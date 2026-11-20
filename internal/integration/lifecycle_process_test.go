package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/discovery"
	"github.com/Hecatoncheir/pob2_engine/internal/host"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
)

type holder struct {
	cmd    *exec.Cmd
	input  io.WriteCloser
	info   pb.GetServerInfoResponse
	done   chan struct{}
	cancel context.CancelFunc
	err    error
}

func launchHolder(ctx context.Context, args ...string) (*holder, error) {
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, binary, append([]string{"connect"}, args...)...)
	cmd.Stderr = os.Stderr
	input, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		input.Close()
		cancel()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		input.Close()
		output.Close()
		cancel()
		return nil, err
	}
	h := &holder{cmd: cmd, input: input, done: make(chan struct{}), cancel: cancel}
	go func() { h.err = cmd.Wait(); close(h.done) }()
	ready := make(chan error, 1)
	go func() { ready <- json.NewDecoder(output).Decode(&h.info) }()
	select {
	case err := <-ready:
		if err == nil {
			return h, nil
		}
		h.cleanup()
		return nil, err
	case <-ctx.Done():
		h.cleanup()
		return nil, ctx.Err()
	}
}

func (h *holder) cleanup() { h.input.Close(); h.cancel(); <-h.done }

func hold(t *testing.T, ctx context.Context, args ...string) *holder {
	t.Helper()
	h, err := launchHolder(ctx, args...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.cleanup)
	return h
}

func stopHolder(t *testing.T, ctx context.Context, h *holder) {
	t.Helper()
	h.input.Close()
	select {
	case <-h.done:
		if h.err != nil {
			t.Fatalf("клиент завершился с ошибкой: %v", h.err)
		}
	case <-ctx.Done():
		t.Fatal("клиент не освободил аренду")
	}
}

func profileArgs(base, name string) []string {
	return []string{"--profile", name, "--runtime-dir", base, "--data-dir", base, "--timeout", "8s"}
}

func autoArgs(base, name string) []string {
	return append(profileArgs(base, name), "--start-server", "--idle-timeout", "300ms", "--lease-timeout", "900ms")
}

func testProfile(t *testing.T, base, name string) discovery.Profile {
	t.Helper()
	p, err := discovery.Open(base, name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		record, err := p.Read()
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		client, err := grpcapi.Dial(ctx, record.Address, record.Token)
		if err == nil {
			_ = client.Shutdown(ctx)
			_ = client.Close()
		}
		for ctx.Err() == nil {
			if _, err := os.Stat(p.RecordPath()); errors.Is(err, os.ErrNotExist) {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
	})
	return p
}

func await(t *testing.T, ctx context.Context, description string, condition func() bool) {
	t.Helper()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		if condition() {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("не выполнено условие: %s", description)
		case <-ticker.C:
		}
	}
}

func inspectProfile(p discovery.Profile) (*pb.GetServerInfoResponse, error) {
	record, err := p.Read()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	return grpcapi.Inspect(ctx, record.Address, record.Token)
}

func awaitGone(t *testing.T, ctx context.Context, p discovery.Profile) {
	t.Helper()
	await(t, ctx, "удаление файла подключения и освобождение блокировки", func() bool {
		if _, err := os.Stat(p.RecordPath()); !errors.Is(err, os.ErrNotExist) {
			return false
		}
		lock, err := p.TryLock("server")
		if err != nil {
			return false
		}
		lock.Close()
		return true
	})
}

func TestAutomaticProfileKeepsSecondClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	base := t.TempDir()
	p := testProfile(t, base, "shared")
	a := hold(t, ctx, append(autoArgs(base, "shared"), "--kind", "ui")...)
	b := hold(t, ctx, append(profileArgs(base, "shared"), "--kind", "ui")...)
	if a.info.GetInstanceId() != b.info.GetInstanceId() {
		t.Fatal("клиенты подключились к разным серверам")
	}
	cli(t, ctx, 0, append([]string{"session", "create", "--session", "main", "--name", "Общая", "--request-id", "create"}, profileArgs(base, "shared")...)...)
	stopHolder(t, ctx, a)
	await(t, ctx, "один оставшийся клиент", func() bool { info, err := inspectProfile(p); return err == nil && info.GetClientCount() == 1 })
	time.Sleep(400 * time.Millisecond)
	cli(t, ctx, 0, append([]string{"session", "get", "--session", "main"}, profileArgs(base, "shared")...)...)
	info, err := inspectProfile(p)
	if err != nil || info.GetInstanceId() != b.info.GetInstanceId() {
		t.Fatalf("сервер второго клиента потерян: %v, %v", info, err)
	}
	stopHolder(t, ctx, b)
	awaitGone(t, ctx, p)
}

func TestConcurrentProfileStartUsesOneServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	base := t.TempDir()
	p := testProfile(t, base, "concurrent")
	type result struct {
		holder *holder
		err    error
	}
	results := make(chan result, 2)
	for range 2 {
		go func() { h, err := launchHolder(ctx, autoArgs(base, "concurrent")...); results <- result{h, err} }()
	}
	var clients []*holder
	for range 2 {
		value := <-results
		if value.err != nil {
			t.Fatal(value.err)
		}
		clients = append(clients, value.holder)
		t.Cleanup(value.holder.cleanup)
	}
	if clients[0].info.GetInstanceId() != clients[1].info.GetInstanceId() {
		t.Fatal("одновременно запущены два сервера")
	}
	info, err := inspectProfile(p)
	if err != nil || info.GetClientCount() != 2 {
		t.Fatalf("два клиента не зарегистрированы: %v, %v", info, err)
	}
	for _, h := range clients {
		stopHolder(t, ctx, h)
	}
	awaitGone(t, ctx, p)
}

func TestCrashedClientLeaseExpires(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	base := t.TempDir()
	p := testProfile(t, base, "crash")
	h := hold(t, ctx, autoArgs(base, "crash")...)
	if err := h.cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	<-h.done
	awaitGone(t, ctx, p)
}

func TestPersistentProfileSurvivesLastClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	base := t.TempDir()
	p := testProfile(t, base, "persistent")
	server := startServer(t, ctx, "--profile", "persistent", "--runtime-dir", base, "--mode", "persistent", "--idle-timeout", "100ms")
	h := hold(t, ctx, autoArgs(base, "persistent")...)
	stopHolder(t, ctx, h)
	time.Sleep(250 * time.Millisecond)
	info, err := inspectProfile(p)
	if err != nil || info.GetMode() != lifecycle.Persistent || info.GetClientCount() != 0 {
		t.Fatalf("постоянный сервер завершился: %v, %v", info, err)
	}
	cli(t, ctx, 0, append([]string{"shutdown"}, profileArgs(base, "persistent")...)...)
	select {
	case <-server.done:
		if server.err != nil {
			t.Fatal(server.err)
		}
	case <-ctx.Done():
		t.Fatal("постоянный сервер не остановился")
	}
	awaitGone(t, ctx, p)
}

func TestStaleRecordDoesNotTrustPID(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	base := t.TempDir()
	p := testProfile(t, base, "stale")
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	if err := p.Publish(discovery.Record{InstanceID: "old-instance", Address: address, Token: "old-token", ProtocolVersion: "pob2.v1", Mode: lifecycle.Automatic, PID: os.Getpid()}); err != nil {
		t.Fatal(err)
	}
	h := hold(t, ctx, autoArgs(base, "stale")...)
	if h.info.GetInstanceId() == "old-instance" {
		t.Fatal("принят устаревший экземпляр по существующему PID")
	}
	stopHolder(t, ctx, h)
	awaitGone(t, ctx, p)
}

func TestExternalMCPKeepsAutomaticServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	base := t.TempDir()
	p := testProfile(t, base, "mcp")
	a := hold(t, ctx, append(autoArgs(base, "mcp"), "--kind", "ui")...)
	cli(t, ctx, 0, append([]string{"session", "create", "--session", "main", "--name", "MCP", "--request-id", "create"}, profileArgs(base, "mcp")...)...)
	m := connectMCP(t, ctx, append([]string{"--attach", "main"}, profileArgs(base, "mcp")...)...)
	stopHolder(t, ctx, a)
	time.Sleep(400 * time.Millisecond)
	result := resultSnapshot(t, call(t, ctx, m, "get_snapshot", map[string]any{}, false))
	if result.BuildName != "MCP" {
		t.Fatalf("MCP потерял сборку: %v", result)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	awaitGone(t, ctx, p)
}

func TestStaleRecordWithReusedPortPreservesOtherServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	foreign := startServer(t, ctx)
	base := t.TempDir()
	p := testProfile(t, base, "reused-port")
	if err := p.Publish(discovery.Record{InstanceID: "stale", Address: foreign.ready.Address, Token: "stale-token", ProtocolVersion: "pob2.v1", Mode: lifecycle.Automatic, PID: os.Getpid()}); err != nil {
		t.Fatal(err)
	}
	h := hold(t, ctx, autoArgs(base, "reused-port")...)
	if h.info.GetInstanceId() == foreign.ready.InstanceID {
		t.Fatal("принят посторонний сервер")
	}
	info, err := grpcapi.Inspect(ctx, foreign.ready.Address, foreign.ready.Token)
	if err != nil || info.GetState() != "ready" {
		t.Fatalf("посторонний сервер затронут: %v, %v", info, err)
	}
	stopHolder(t, ctx, h)
	awaitGone(t, ctx, p)
}

func TestLifecycleHelper(t *testing.T) {
	if os.Getenv("POB2_LIFECYCLE_HELPER") != "1" {
		t.Skip("вспомогательный сервер отдельного процесса")
	}
	base := os.Getenv("POB2_HELPER_BASE")
	profile, err := discovery.Open(base, "task")
	if err != nil {
		t.Fatal(err)
	}
	s, err := host.StartDurable(context.Background(), profile, "127.0.0.1:0", "", base, lifecycle.Config{Mode: lifecycle.Automatic, IdleTimeout: 300 * time.Millisecond, LeaseTimeout: 900 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = s.Close(ctx)
	}()
	finish, err := s.API.Lifecycle().BeginTask("")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	go func() {
		defer finish()
		timer := time.NewTicker(20 * time.Millisecond)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.API.Lifecycle().Context().Done():
				return
			case <-timer.C:
				if _, err := os.Stat(filepath.Join(base, "finish-task")); err == nil {
					return
				}
			}
		}
	}()
	if err := json.NewEncoder(os.Stdout).Encode(s.Ready()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-s.Done():
		if err := s.Err(); err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("задача вспомогательного сервера не завершилась")
	}
}

func startTaskServer(t *testing.T, ctx context.Context, base string) <-chan error {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, executable, "-test.run=^TestLifecycleHelper$")
	cmd.Env = append(os.Environ(), "POB2_LIFECYCLE_HELPER=1", "POB2_HELPER_BASE="+base)
	cmd.Stderr = os.Stderr
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- cmd.Wait() }()
	ready := make(chan error, 1)
	go func() { var value grpcapi.Ready; ready <- json.NewDecoder(output).Decode(&value) }()
	select {
	case err := <-ready:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("нет готовности вспомогательного процесса")
	}
	return result
}

func TestBackgroundTaskKeepsAutomaticProcess(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	base := t.TempDir()
	p := testProfile(t, base, "task")
	result := startTaskServer(t, ctx, base)
	h := hold(t, ctx, profileArgs(base, "task")...)
	stopHolder(t, ctx, h)
	time.Sleep(400 * time.Millisecond)
	info, err := inspectProfile(p)
	if err != nil || info.GetClientCount() != 0 || info.GetActiveTasks() != 1 {
		t.Fatalf("задача не удерживает процесс: %v, %v", info, err)
	}
	if err := os.WriteFile(filepath.Join(base, "finish-task"), []byte("done"), 0600); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("процесс не завершился после фоновой задачи")
	}
	awaitGone(t, ctx, p)
}

func TestDiscoveryWaitsForDrainingInstance(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	base := t.TempDir()
	p := testProfile(t, base, "task")
	oldProcess := startTaskServer(t, ctx, base)
	old, err := p.Read()
	if err != nil {
		t.Fatal(err)
	}
	cli(t, ctx, 0, append([]string{"shutdown"}, profileArgs(base, "task")...)...)
	type connected struct {
		value *holder
		err   error
	}
	ready := make(chan connected, 1)
	go func() { h, err := launchHolder(ctx, autoArgs(base, "task")...); ready <- connected{h, err} }()
	select {
	case value := <-ready:
		if value.value != nil {
			value.value.cleanup()
		}
		t.Fatalf("подключение завершилось до остановки старого процесса: %v", value.err)
	case <-time.After(150 * time.Millisecond):
	}
	if record, err := p.Read(); err != nil || record.InstanceID != old.InstanceID {
		t.Fatalf("файл старого процесса заменён преждевременно: %v, %v", record, err)
	}
	if err := os.WriteFile(filepath.Join(base, "finish-task"), []byte("done"), 0600); err != nil {
		t.Fatal(err)
	}
	var next *holder
	select {
	case value := <-ready:
		if value.err != nil {
			t.Fatal(value.err)
		}
		next = value.value
	case <-ctx.Done():
		t.Fatal("новый сервер не запустился")
	}
	t.Cleanup(next.cleanup)
	if next.info.GetInstanceId() == old.InstanceID {
		t.Fatal("повторно принят завершающийся сервер")
	}
	select {
	case err := <-oldProcess:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("старый процесс не завершился")
	}
	stopHolder(t, ctx, next)
	awaitGone(t, ctx, p)
}
