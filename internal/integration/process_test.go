package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var binary string

func TestMain(m *testing.M) {
	if os.Getenv("POB2_LIFECYCLE_HELPER") == "1" {
		os.Exit(m.Run())
	}
	dir, err := os.MkdirTemp("", "pob2-process-test-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	binary = filepath.Join(dir, "pob2-engine")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	cmd := exec.CommandContext(ctx, "go", "build", "-ldflags=-X github.com/Hecatoncheir/pob2_engine/internal/buildinfo.Version=integration-test", "-o", binary, "./cmd/pob2-engine")
	cmd.Dir = filepath.Join("..", "..")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err = cmd.Run()
	cancel()
	code := 1
	if err == nil {
		code = m.Run()
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

type process struct {
	ready grpcapi.Ready
	done  chan struct{}
	err   error
}

func startServer(t *testing.T, ctx context.Context, args ...string) *process {
	t.Helper()
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, binary, append([]string{"serve"}, args...)...)
	// Тест изолирован от пользовательской настройки токена.
	cmd.Env = append(os.Environ(), "POB2_ENGINE_TOKEN=")
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	p := &process{done: make(chan struct{})}
	go func() { p.err = cmd.Wait(); close(p.done) }()
	t.Cleanup(func() { cancel(); <-p.done })
	ready := make(chan error, 1)
	go func() { ready <- json.NewDecoder(stdout).Decode(&p.ready) }()
	select {
	case err := <-ready:
		if err != nil {
			t.Fatalf("готовность сервера: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("сервер не сообщил о готовности")
	}
	return p
}

func cli(t *testing.T, ctx context.Context, wantExit int, args ...string) []byte {
	t.Helper()
	cmd := exec.CommandContext(ctx, binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != wantExit {
		t.Fatalf("CLI: %v; stderr: %s", err, &stderr)
	}
	if wantExit != 0 && stdout.Len() != 0 {
		t.Fatalf("ошибка попала в stdout: %s", &stdout)
	}
	return stdout.Bytes()
}

func connectMCP(t *testing.T, ctx context.Context, args ...string) *mcp.ClientSession {
	t.Helper()
	client := mcp.NewClient(&mcp.Implementation{Name: "process-test", Version: "1"}, nil)
	cmd := exec.CommandContext(ctx, binary, append([]string{"mcp"}, args...)...)
	cmd.Stderr = os.Stderr
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd, TerminateDuration: time.Second}, nil)
	if err != nil {
		t.Fatalf("MCP initialize: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 2 {
		t.Fatalf("MCP tools/list: %v, %v", tools, err)
	}
	return session
}

func call(t *testing.T, ctx context.Context, session *mcp.ClientSession, name string, args map[string]any, wantError bool) *mcp.CallToolResult {
	t.Helper()
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError != wantError {
		t.Fatalf("MCP %s: %+v", name, result)
	}
	return result
}

func resultSnapshot(t *testing.T, result *mcp.CallToolResult) domain.Snapshot {
	t.Helper()
	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var value domain.Snapshot
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func TestCLIAndMCPShareSessionAcrossProcesses(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	p := startServer(t, ctx)
	connection := []string{"--address", p.ready.Address, "--token", p.ready.Token}
	create := append([]string{"session", "create", "--session", "main", "--name", "Исходная", "--request-id", "create-1"}, connection...)
	var initial domain.Snapshot
	if err := json.Unmarshal(cli(t, ctx, 0, create...), &initial); err != nil {
		t.Fatal(err)
	}
	if initial.Revision != 1 {
		t.Fatalf("начальная ревизия: %v", initial)
	}
	m := connectMCP(t, ctx, append([]string{"--attach", "main"}, connection...)...)
	input := map[string]any{"build_name": "Из MCP", "expected_revision": 1, "request_id": "rename-1"}
	changed := resultSnapshot(t, call(t, ctx, m, "rename_build", input, false))
	if changed.BuildName != "Из MCP" || changed.Revision != 2 {
		t.Fatalf("правка MCP: %v", changed)
	}
	call(t, ctx, m, "rename_build", map[string]any{"build_name": "Конфликт", "expected_revision": 1, "request_id": "stale"}, true)
	call(t, ctx, m, "rename_build", map[string]any{"build_name": "Подмена", "expected_revision": 1, "request_id": "rename-1"}, true)
	latest := resultSnapshot(t, call(t, ctx, m, "rename_build", map[string]any{"build_name": "Последняя", "expected_revision": 2, "request_id": "rename-2"}, false))
	replay := resultSnapshot(t, call(t, ctx, m, "rename_build", input, false))
	if replay != changed {
		t.Fatalf("повтор не вернул исходный результат: %v", replay)
	}
	if got := resultSnapshot(t, call(t, ctx, m, "get_snapshot", map[string]any{}, false)); got != latest {
		t.Fatalf("снимок MCP: %v", got)
	}
	if err := m.Close(); err != nil {
		t.Fatalf("завершение MCP: %v", err)
	}
	var fromCLI domain.Snapshot
	get := append([]string{"session", "get", "--session", "main"}, connection...)
	if err := json.Unmarshal(cli(t, ctx, 0, get...), &fromCLI); err != nil {
		t.Fatal(err)
	}
	if fromCLI != latest || fromCLI.Revision != 3 {
		t.Fatalf("CLI видит другое состояние: %v", fromCLI)
	}
	cli(t, ctx, 1, append([]string{"mcp", "--attach", "missing"}, connection...)...)
	cli(t, ctx, 1, append([]string{"session", "get", "--session", "missing"}, connection...)...)
	cli(t, ctx, 0, append([]string{"shutdown"}, connection...)...)
	select {
	case <-p.done:
		if p.err != nil {
			t.Fatalf("сервер завершился с ошибкой: %v", p.err)
		}
	case <-ctx.Done():
		t.Fatal("сервер завис при завершении")
	}
}

func TestStandaloneMCPAndExplicitMode(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cli(t, ctx, 2, "mcp")
	cli(t, ctx, 2, "mcp", "--standalone", "--attach", "main")
	m := connectMCP(t, ctx, "--standalone", "--session", "private", "--name", "Локальная")
	snapshot := resultSnapshot(t, call(t, ctx, m, "get_snapshot", map[string]any{}, false))
	if snapshot.SessionID != "private" || snapshot.BuildName != "Локальная" || snapshot.Revision != 1 {
		t.Fatalf("standalone: %v", snapshot)
	}
	if err := m.Close(); err != nil {
		t.Fatalf("MCP не завершился после EOF: %v", err)
	}
}

func TestCLIHelpVersionAndExitCodes(t *testing.T) {
	if !strings.Contains(string(cli(t, t.Context(), 0, "--help")), "session") {
		t.Fatal("нет справки")
	}
	var version map[string]string
	if err := json.Unmarshal(cli(t, t.Context(), 0, "--version"), &version); err != nil {
		t.Fatal(err)
	}
	if version["protocol_version"] != "pob2.v1" || version["version"] != "integration-test" {
		t.Fatalf("версия: %v", version)
	}
	cli(t, t.Context(), 2, "unknown")
	cli(t, t.Context(), 2, "serve", "--unknown")
	cli(t, t.Context(), 1, "serve", "--listen", "0.0.0.0:0")
}
