package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type child struct {
	cmd   *exec.Cmd
	input io.WriteCloser
	done  chan error
}

func startChild(ctx context.Context, target target, scratch string, ready any, args ...string) (*child, error) {
	cmd := exec.CommandContext(ctx, target.Binary, args...)
	cmd.Env = append(os.Environ(), "POB2_ENGINE_POB_DIR="+target.PoB,
		"POB2_ENGINE_POB_CACHE_DIR="+filepath.Join(scratch, "pob-cache"), "POB2_ENGINE_CALCULATOR=lua")
	cmd.Stderr = os.Stderr
	input, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		_ = input.Close()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	process := &child{cmd: cmd, input: input, done: make(chan error, 1)}
	err = json.NewDecoder(output).Decode(ready)
	go func() { process.done <- cmd.Wait() }()
	if err != nil {
		_ = cmd.Process.Kill()
		<-process.done
		return nil, fmt.Errorf("готовность %s: %w", args[0], err)
	}
	return process, nil
}

func (c *child) close() {
	_ = c.input.Close()
	timer := time.NewTimer(6 * time.Second)
	defer timer.Stop()
	select {
	case <-c.done:
	case <-timer.C:
		_ = c.cmd.Process.Kill()
		<-c.done
	}
}

func stopSession(client *grpcapi.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = client.Shutdown(ctx)
	_ = client.Close()
}

type workbenchClient struct {
	connection *grpc.ClientConn
	rpc        pb.WorkbenchServiceClient
	token      string
}

func openWorkbench(ready grpcapi.WorkbenchReady) (*workbenchClient, error) {
	if err := grpcapi.ValidateTarget(ready.Address, ready.Token); err != nil {
		return nil, err
	}
	channel, err := grpc.NewClient("passthrough:///"+ready.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &workbenchClient{connection: channel, rpc: pb.NewWorkbenchServiceClient(channel), token: ready.Token}, nil
}

func (c *workbenchClient) importXML(ctx context.Context, xml string) error {
	input, err := json.Marshal(map[string]string{"xml": xml})
	if err != nil {
		return err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
	reply, err := c.rpc.CallWorkbenchTool(ctx, &pb.CallWorkbenchToolRequest{Name: "pob_import_xml", ArgumentsJson: input})
	if err != nil {
		return err
	}
	if failure := reply.GetError(); failure != nil {
		return fmt.Errorf("%s: %s", failure.GetCode(), failure.GetMessage())
	}
	if reply.GetRevision() == 0 || len(reply.GetResultJson()) == 0 {
		return fmt.Errorf("рабочая копия не подтвердила импорт XML")
	}
	return nil
}
