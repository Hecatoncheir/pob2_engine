package grpcapi

import (
	"context"
	"testing"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestEditorBridgeChecksWindowLease(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 8*time.Second)
	defer cancel()
	service := application.NewService()
	if _, err := service.Create(ctx, application.CreateSession{SessionID: "main", BuildName: "Сборка", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	server, err := Start("127.0.0.1:0", "", service)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Close(ctx)
	})
	tool := connectEditorTest(t, ctx, server, "mcp")
	if _, err := tool.CallEditor(ctx, "main", "pob_editor_status", "{}"); status.Code(err) != codes.Unavailable {
		t.Fatal("команда без разрешённого окна принята", err)
	}
	window := connectEditorTest(t, ctx, server, "ui")
	watch, err := window.rpc.WatchEditor(ctx, &pb.WatchEditorRequest{SessionId: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if attached, err := watch.Recv(); err != nil || !attached.GetAttached() {
		t.Fatal("не подтверждена привязка редактора", err)
	}
	results := make(chan error, 1)
	go func() {
		result, err := tool.CallEditor(ctx, "main", "pob_editor_ui", `{"action":"snapshot"}`)
		if err == nil && result != `{"result":{"dialogOpen":false}}` {
			err = status.Errorf(codes.DataLoss, "ответ изменился: %q", result)
		}
		results <- err
	}()
	command, err := watch.Recv()
	if err != nil || command.GetTool() != "pob_editor_ui" || command.GetArgumentsJson() != `{"action":"snapshot"}` {
		t.Fatal("поток не передал команду", command, err)
	}
	reply := &pb.ReplyEditorRequest{SessionId: "main", RequestId: command.GetRequestId(), ResultJson: `{"result":{"dialogOpen":false}}`}
	if _, err := tool.rpc.ReplyEditor(ctx, reply); status.Code(err) != codes.PermissionDenied {
		t.Fatal("ответ посторонней аренды принят", err)
	}
	if _, err := window.rpc.ReplyEditor(ctx, reply); err != nil {
		t.Fatal(err)
	}
	if err := <-results; err != nil {
		t.Fatal(err)
	}
}

func connectEditorTest(t *testing.T, ctx context.Context, server *Server, kind string) *Client {
	t.Helper()
	ready := server.Ready()
	client, err := DialKind(ctx, ready.Address, ready.Token, kind)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}
