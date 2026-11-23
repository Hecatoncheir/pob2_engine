package grpcapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"slices"
	"sync"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const editorArgumentsLimit = 1 << 20

var editorTools = []string{
	"pob_editor_status", "pob_editor_app", "pob_editor_components",
	"pob_editor_library", "pob_editor_history", "pob_editor_notes",
	"pob_editor_files", "pob_editor_progression", "pob_editor_ui", "pob_editor_job",
}

type editorBridge struct {
	mu    sync.Mutex
	links map[string]*editorLink
}

type editorLink struct {
	lease   string
	out     chan *editorCall
	done    chan struct{}
	pending *editorCall
}

type editorCall struct {
	ctx      context.Context
	command  *pb.WatchEditorResponse
	result   chan string
	answered bool
}

func (b *editorBridge) attach(id, lease string) (*editorLink, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.links[id] != nil {
		return nil, status.Error(codes.AlreadyExists, "к сессии уже подключено окно")
	}
	if b.links == nil {
		b.links = make(map[string]*editorLink)
	}
	link := &editorLink{lease: lease, out: make(chan *editorCall, 1), done: make(chan struct{})}
	b.links[id] = link
	return link, nil
}

func (b *editorBridge) detach(id string, link *editorLink) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.links, id)
	close(link.done)
}

func (s *Server) WatchEditor(req *pb.WatchEditorRequest, stream grpc.ServerStreamingServer[pb.WatchEditorResponse]) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	session := req.GetSessionId()
	if _, err := s.commands.Snapshot(ctx, session); err != nil {
		return rpcError(err)
	}
	lease := editorLease(ctx)
	link, err := s.editors.attach(session, lease)
	if err != nil {
		return err
	}
	defer s.editors.detach(session, link)
	go s.watchLease(ctx, cancel, lease, true)
	return sendEditorCommands(ctx, link, stream)
}

func sendEditorCommands(ctx context.Context, link *editorLink, stream grpc.ServerStreamingServer[pb.WatchEditorResponse]) error {
	if err := stream.Send(&pb.WatchEditorResponse{Attached: true}); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return rpcError(ctx.Err())
		case call := <-link.out:
			if call.ctx.Err() == nil {
				if err := stream.Send(call.command); err != nil {
					return err
				}
			}
		}
	}
}

func editorLease(ctx context.Context) string {
	values := metadata.ValueFromIncomingContext(ctx, "pob2-lease")
	if len(values) != 1 {
		return ""
	}
	return values[0]
}

func (s *Server) CallEditor(ctx context.Context, req *pb.CallEditorRequest) (*pb.CallEditorResponse, error) {
	if !slices.Contains(editorTools, req.GetTool()) || !editorArguments(req.GetArgumentsJson()) {
		return nil, status.Error(codes.InvalidArgument, "неверная команда редактора")
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	id := req.GetSessionId()
	call := &editorCall{ctx: ctx, result: make(chan string, 1), command: &pb.WatchEditorResponse{
		RequestId: rand.Text(), Tool: req.GetTool(), ArgumentsJson: req.GetArgumentsJson(),
	}}
	link, err := s.editors.submit(id, call)
	if err != nil {
		return nil, err
	}
	defer s.editors.clear(link, call)
	select {
	case <-ctx.Done():
		return nil, rpcError(ctx.Err())
	case <-link.done:
		select {
		case result := <-call.result:
			return &pb.CallEditorResponse{ResultJson: result}, nil
		default:
		}
		return nil, status.Error(codes.Unavailable, "редактор PoB II отключился")
	case result := <-call.result:
		return &pb.CallEditorResponse{ResultJson: result}, nil
	}
}

func editorArguments(raw string) bool {
	var object map[string]json.RawMessage
	return len(raw) <= editorArgumentsLimit && json.Unmarshal([]byte(raw), &object) == nil && object != nil
}

func (b *editorBridge) submit(id string, call *editorCall) (*editorLink, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	link := b.links[id]
	if link == nil {
		return nil, status.Error(codes.Unavailable, "редактор PoB II не подключён или доступ отключён")
	}
	if link.pending != nil {
		return nil, status.Error(codes.ResourceExhausted, "выполняется другая команда редактора")
	}
	select {
	case link.out <- call:
		link.pending = call
		return link, nil
	default:
		return nil, status.Error(codes.ResourceExhausted, "редактор PoB II занят")
	}
}

func (b *editorBridge) clear(link *editorLink, call *editorCall) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if link.pending == call {
		link.pending = nil
	}
}

func (s *Server) ReplyEditor(ctx context.Context, req *pb.ReplyEditorRequest) (*pb.ReplyEditorResponse, error) {
	if !json.Valid([]byte(req.GetResultJson())) {
		return nil, status.Error(codes.InvalidArgument, "неверный ответ редактора")
	}
	if err := s.editors.reply(editorLease(ctx), req); err != nil {
		return nil, err
	}
	return &pb.ReplyEditorResponse{}, nil
}

func (b *editorBridge) reply(lease string, req *pb.ReplyEditorRequest) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	link := b.links[req.GetSessionId()]
	if link == nil || lease == "" || link.lease != lease {
		return status.Error(codes.PermissionDenied, "ответ должен отправить подключённый редактор")
	}
	call := link.pending
	if call == nil || call.command.RequestId != req.GetRequestId() || call.ctx.Err() != nil {
		return status.Error(codes.NotFound, "редактор ответил на неактуальную команду")
	}
	if call.answered {
		return status.Error(codes.AlreadyExists, "редактор уже ответил")
	}
	call.answered = true
	call.result <- req.GetResultJson()
	return nil
}

// CallEditor передаёт команду в окно выбранной сессии.
func (c *Client) CallEditor(ctx context.Context, sessionID, tool, arguments string) (string, error) {
	reply, err := c.rpc.CallEditor(ctx, &pb.CallEditorRequest{
		SessionId: sessionID, Tool: tool, ArgumentsJson: arguments,
	})
	if err != nil {
		return "", err
	}
	return reply.GetResultJson(), nil
}
