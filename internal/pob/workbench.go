package pob

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/Hecatoncheir/pob2_engine/internal/poblegacy"
)

type Workbench struct {
	process  *worker
	gate     chan struct{}
	close    sync.Once
	sequence int
	revision int
}

type WorkbenchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *WorkbenchError) Error() string { return e.Code + ": " + e.Message }

type WorkbenchResult struct {
	Revision int             `json:"revision"`
	Result   json.RawMessage `json:"result"`
}

func NewWorkbench(ctx context.Context, installation Installation) (*Workbench, error) {
	layout, err := installation.Layout()
	if err != nil {
		return nil, err
	}
	worker, err := startLuaProcess(layout, "_workbench-lua", unpackWorkbench)
	if err != nil {
		return nil, err
	}
	session := &Workbench{process: worker, gate: make(chan struct{}, 1)}
	if _, err := session.Call(ctx, "initialize", nil); err != nil {
		session.Close()
		return nil, err
	}
	return session, nil
}

func (s *Workbench) Close() { s.close.Do(s.process.close) }

func unpackWorkbench(directory string) (string, error) {
	if err := poblegacy.Unpack(directory, "lua"); err != nil {
		return "", err
	}
	if err := poblegacy.Unpack(directory, "data"); err != nil {
		return "", err
	}
	geometry, err := bridge.ReadFile("bridge/tree_geometry.lua")
	if err != nil {
		return "", err
	}
	return directory, os.WriteFile(filepath.Join(directory, "tree_geometry.lua"), geometry, 0600)
}

type workbenchRequest struct {
	Protocol int            `json:"protocol"`
	ID       string         `json:"id"`
	Method   string         `json:"method"`
	Params   map[string]any `json:"params"`
}

type workbenchReply struct {
	Protocol int    `json:"protocol"`
	ID       string `json:"id"`
	WorkbenchResult
	Error *WorkbenchError `json:"error"`
}

type workbenchAnswer struct {
	reply workbenchReply
	err   error
}

func (s *Workbench) Call(ctx context.Context, method string, arguments map[string]any) (WorkbenchResult, error) {
	release, err := acquireSlot(ctx, s.gate)
	if err != nil {
		return WorkbenchResult{}, err
	}
	defer release()
	request := s.request(method, arguments)
	answer := make(chan workbenchAnswer, 1)
	go func() {
		reply, err := s.exchange(request)
		answer <- workbenchAnswer{reply, err}
	}()
	select {
	case <-ctx.Done():
		s.Close()
		return WorkbenchResult{}, ctx.Err()
	case value := <-answer:
		return s.accept(value)
	}
}

func (s *Workbench) request(method string, arguments map[string]any) workbenchRequest {
	s.sequence++
	params := make(map[string]any, len(arguments)+1)
	for key, value := range arguments {
		params[key] = value
	}
	params["revision"] = s.revision
	return workbenchRequest{Protocol: 1, ID: strconv.Itoa(s.sequence), Method: method, Params: params}
}

func (s *Workbench) accept(value workbenchAnswer) (WorkbenchResult, error) {
	if value.err != nil {
		s.Close()
		return WorkbenchResult{}, value.err
	}
	s.revision = value.reply.Revision
	if value.reply.Error != nil {
		return WorkbenchResult{}, value.reply.Error
	}
	return value.reply.WorkbenchResult, nil
}

func (s *Workbench) exchange(request workbenchRequest) (workbenchReply, error) {
	var reply workbenchReply
	if err := json.NewEncoder(s.process.input).Encode(request); err != nil {
		return reply, err
	}
	if !s.process.output.Scan() {
		err := s.process.output.Err()
		if err == nil {
			err = io.EOF
		}
		return reply, fmt.Errorf("чтение ответа мастерской PoB: %w", err)
	}
	if err := json.Unmarshal(s.process.output.Bytes(), &reply); err != nil {
		return reply, fmt.Errorf("неверный ответ мастерской PoB: %w", err)
	}
	if reply.Protocol != 1 || reply.ID != request.ID || reply.Revision < 0 {
		return reply, fmt.Errorf("нарушена последовательность ответов мастерской PoB")
	}
	return reply, nil
}
