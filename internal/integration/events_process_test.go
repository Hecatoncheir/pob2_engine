package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
)

func TestCLIAndMCPChangesReachWatchAndResumeAfterRestart(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	runtimeDir, dataDir := t.TempDir(), t.TempDir()
	p := testProfile(t, runtimeDir, "events")
	args := storedArgs(runtimeDir, dataDir, "events")
	h := hold(t, ctx, storedAutoArgs(runtimeDir, dataDir, "events")...)
	commandSnapshot(t, ctx, args, "session", "create", "--session", "main", "--name", "До", "--request-id", "create")
	record, err := p.Read()
	if err != nil {
		t.Fatal(err)
	}
	ui, err := grpcapi.DialKind(ctx, record.Address, record.Token, "ui")
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Close()
	stream, err := ui.Watch(ctx, "main", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stream.Recv(); err != nil {
		t.Fatal(err)
	}
	commandSnapshot(t, ctx, args, "session", "rename", "--session", "main", "--name", "CLI", "--revision", "1", "--request-id", "cli")
	m := connectMCP(t, ctx, append([]string{"--attach", "main"}, args...)...)
	input := map[string]any{"build_name": "MCP", "expected_revision": 2, "request_id": "mcp"}
	call(t, ctx, m, "rename_build", input, false)
	call(t, ctx, m, "rename_build", input, false)
	for i, name := range []string{"CLI", "MCP"} {
		event, err := stream.Recv()
		if err != nil || event.GetSnapshot().GetBuildName() != name || event.GetSnapshot().GetRevision() != uint64(i+2) {
			t.Fatalf("событие: %v %v", event, err)
		}
	}
	ui.Close()
	m.Close()
	stopHolder(t, ctx, h)
	awaitGone(t, ctx, p)
	next := hold(t, ctx, storedAutoArgs(runtimeDir, dataDir, "events")...)
	record, err = p.Read()
	if err != nil {
		t.Fatal(err)
	}
	ui, err = grpcapi.DialKind(ctx, record.Address, record.Token, "ui")
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Close()
	stream, err = ui.Watch(ctx, "main", 2)
	if err != nil {
		t.Fatal(err)
	}
	event, err := stream.Recv()
	if err != nil || event.GetRequestId() != "mcp" || event.GetSnapshot().GetRevision() != 3 {
		t.Fatalf("пропущенное событие: %v %v", event, err)
	}
	ui.Close()
	stopHolder(t, ctx, next)
	awaitGone(t, ctx, p)
}
