package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/discovery"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/host"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/storage"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
)

func storedArgs(runtimeDir, dataDir, name string) []string {
	return []string{"--profile", name, "--runtime-dir", runtimeDir, "--data-dir", dataDir, "--timeout", "8s"}
}

func storedAutoArgs(runtimeDir, dataDir, name string) []string {
	return append(storedArgs(runtimeDir, dataDir, name), "--start-server", "--idle-timeout", "300ms", "--lease-timeout", "900ms")
}

func commandSnapshot(t *testing.T, ctx context.Context, connection []string, args ...string) domain.Snapshot {
	t.Helper()
	var value domain.Snapshot
	if err := json.Unmarshal(cli(t, ctx, 0, append(args, connection...)...), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func TestAutomaticRestartRestoresMCPChangesAndReplay(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	runtimeDir, dataDir := t.TempDir(), t.TempDir()
	p := testProfile(t, runtimeDir, "saved")
	connection := storedArgs(runtimeDir, dataDir, "saved")
	h := hold(t, ctx, storedAutoArgs(runtimeDir, dataDir, "saved")...)
	initial := commandSnapshot(t, ctx, connection, "session", "create", "--session", "main", "--name", "Начальная", "--request-id", "create")
	m := connectMCP(t, ctx, append([]string{"--attach", "main"}, connection...)...)
	input := map[string]any{"build_name": "Сохранена через MCP", "expected_revision": 1, "request_id": "mcp-rename"}
	changed := resultSnapshot(t, call(t, ctx, m, "rename_build", input, false))
	latest := commandSnapshot(t, ctx, connection, "session", "rename", "--session", "main", "--name", "Последняя", "--revision", "2", "--request-id", "cli-rename")
	oldID := h.info.GetInstanceId()
	if h.info.GetStorageMode() != "disk" || h.info.GetStoragePath() == "" {
		t.Fatal("сервер не объявил сохранение")
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	stopHolder(t, ctx, h)
	awaitGone(t, ctx, p)
	if _, err := os.Stat(h.info.GetStoragePath()); err != nil {
		t.Fatal("база удалена вместе со служебным файлом")
	}
	next := hold(t, ctx, storedAutoArgs(runtimeDir, dataDir, "saved")...)
	if next.info.GetInstanceId() == oldID {
		t.Fatal("сервер не перезапускался")
	}
	got := commandSnapshot(t, ctx, connection, "session", "get", "--session", "main")
	if got != latest {
		t.Fatalf("сессия не восстановлена: %v, нужно %v", got, latest)
	}
	createdAgain := commandSnapshot(t, ctx, connection, "session", "create", "--session", "main", "--name", "Начальная", "--request-id", "create")
	if createdAgain != initial {
		t.Fatalf("не восстановлен результат создания: %v", createdAgain)
	}
	m = connectMCP(t, ctx, append([]string{"--attach", "main"}, connection...)...)
	if replay := resultSnapshot(t, call(t, ctx, m, "rename_build", input, false)); replay != changed {
		t.Fatalf("правка повторно применена после запуска: %v", replay)
	}
	if got := resultSnapshot(t, call(t, ctx, m, "get_snapshot", map[string]any{}, false)); got != latest {
		t.Fatalf("повтор изменил текущую ревизию: %v", got)
	}
	m.Close()
	stopHolder(t, ctx, next)
	awaitGone(t, ctx, p)
}

func TestKilledServerRestoresAcknowledgedCommandsWithoutRuntimeCache(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	runtimeDir, dataDir := t.TempDir(), t.TempDir()
	testProfile(t, runtimeDir, "crash-saved")
	s := startServer(t, ctx, "--profile", "crash-saved", "--runtime-dir", runtimeDir, "--data-dir", dataDir)
	connection := storedArgs(runtimeDir, dataDir, "crash-saved")
	commandSnapshot(t, ctx, connection, "session", "create", "--session", "main", "--name", "Начальная", "--request-id", "create")
	last := commandSnapshot(t, ctx, connection, "session", "rename", "--session", "main", "--name", "Подтверждена", "--revision", "1", "--request-id", "rename")
	if err := s.cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	<-s.done
	newRuntime := t.TempDir()
	p := testProfile(t, newRuntime, "crash-saved")
	next := hold(t, ctx, storedAutoArgs(newRuntime, dataDir, "crash-saved")...)
	connection = storedArgs(newRuntime, dataDir, "crash-saved")
	if got := commandSnapshot(t, ctx, connection, "session", "get", "--session", "main"); got != last {
		t.Fatalf("подтверждённая правка потеряна после Kill: %v", got)
	}
	if got := commandSnapshot(t, ctx, connection, "session", "rename", "--session", "main", "--name", "Подтверждена", "--revision", "1", "--request-id", "rename"); got != last {
		t.Fatalf("не восстановлен request_id: %v", got)
	}
	if next.info.GetActiveTasks() != 0 {
		t.Fatal("задачи предыдущего процесса восстановлены как активные")
	}
	stopHolder(t, ctx, next)
	awaitGone(t, ctx, p)
}

func TestSecondRuntimeCannotWriteSameStorage(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	runtimeDir, otherRuntime, dataDir := t.TempDir(), t.TempDir(), t.TempDir()
	p := testProfile(t, runtimeDir, "shared-data")
	testProfile(t, otherRuntime, "shared-data")
	h := hold(t, ctx, storedAutoArgs(runtimeDir, dataDir, "shared-data")...)
	connection := storedArgs(runtimeDir, dataDir, "shared-data")
	initial := commandSnapshot(t, ctx, connection, "session", "create", "--session", "main", "--name", "Сохранённая", "--request-id", "create")
	cli(t, ctx, 1, "serve", "--profile", "shared-data", "--runtime-dir", otherRuntime, "--data-dir", dataDir)
	if got := commandSnapshot(t, ctx, connection, "session", "get", "--session", "main"); got != initial {
		t.Fatal("второй сервер изменил данные")
	}
	stopHolder(t, ctx, h)
	awaitGone(t, ctx, p)
}

func TestDataDirectoryMismatchDoesNotSwitchSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	runtimeDir, dataDir, wrongData := t.TempDir(), t.TempDir(), t.TempDir()
	p := testProfile(t, runtimeDir, "data-path")
	h := hold(t, ctx, storedAutoArgs(runtimeDir, dataDir, "data-path")...)
	cli(t, ctx, 1, append([]string{"info"}, storedArgs(runtimeDir, wrongData, "data-path")...)...)
	if _, err := os.Stat(filepath.Join(wrongData, "pob2-engine")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("создано второе хранилище вместо ошибки")
	}
	info, err := inspectProfile(p)
	if err != nil || info.GetInstanceId() != h.info.GetInstanceId() {
		t.Fatal("запущен другой экземпляр")
	}
	stopHolder(t, ctx, h)
	awaitGone(t, ctx, p)
}

func TestCorruptStorageStopsStartupWithoutReset(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	runtimeDir, dataDir := t.TempDir(), t.TempDir()
	p := testProfile(t, runtimeDir, "damaged")
	path, err := storage.Path(dataDir, "damaged")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	original := []byte("повреждённое хранилище")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	cli(t, ctx, 1, append([]string{"connect"}, storedAutoArgs(runtimeDir, dataDir, "damaged")...)...)
	if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, original) {
		t.Fatal("повреждённый файл сброшен или перезаписан")
	}
	if _, err := p.Read(); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("опубликована готовность без восстановления данных")
	}
}

func TestDefaultServeIsDurableAndEphemeralIsExplicit(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	runtimeDir, dataDir := t.TempDir(), t.TempDir()
	testProfile(t, runtimeDir, "default")
	s := startServer(t, ctx, "--runtime-dir", runtimeDir, "--data-dir", dataDir)
	if s.ready.StorageMode != "disk" {
		t.Fatal("обычный serve не сохраняет данные")
	}
	cli(t, ctx, 0, "shutdown", "--address", s.ready.Address, "--token", s.ready.Token)
	<-s.done
	cli(t, ctx, 2, "serve", "--ephemeral", "--data-dir", dataDir)
	ephemeral := startServer(t, ctx, "--ephemeral")
	if ephemeral.ready.StorageMode != "memory" || ephemeral.ready.StoragePath != "" {
		t.Fatal("ephemeral объявляет постоянное хранилище")
	}
	cli(t, ctx, 0, "shutdown", "--address", ephemeral.ready.Address, "--token", ephemeral.ready.Token)
	<-ephemeral.done
}

type failingJournal struct {
	*storage.Journal
	failurePath string
}

func (j failingJournal) Append(ctx context.Context, change application.Change) error {
	if _, err := os.Stat(j.failurePath); err == nil {
		return errors.New("тестовый отказ записи на диск")
	}
	return j.Journal.Append(ctx, change)
}

func TestPersistenceFailureHelper(t *testing.T) {
	if os.Getenv("POB2_PERSISTENCE_HELPER") != "1" {
		t.Skip("вспомогательный процесс отказа записи")
	}
	base := os.Getenv("POB2_HELPER_BASE")
	profile, err := discovery.Open(base, "io-error")
	if err != nil {
		t.Fatal(err)
	}
	journal, err := storage.Open(t.Context(), base, "io-error")
	if err != nil {
		t.Fatal(err)
	}
	defer journal.Close()
	service, err := application.OpenService(t.Context(), failingJournal{Journal: journal, failurePath: filepath.Join(base, "fail-save")})
	if err != nil {
		t.Fatal(err)
	}
	s, err := host.Start(profile, "127.0.0.1:0", "", service, lifecycle.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = s.Close(ctx)
	}()
	if err := json.NewEncoder(os.Stdout).Encode(s.Ready()); err != nil {
		t.Fatal(err)
	}
	<-s.Done()
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
}

func startFailureServer(t *testing.T, ctx context.Context, base string) <-chan error {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, executable, "-test.run=^TestPersistenceFailureHelper$")
	cmd.Env = append(os.Environ(), "POB2_PERSISTENCE_HELPER=1", "POB2_HELPER_BASE="+base)
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
	go func() { var info grpcapi.Ready; ready <- json.NewDecoder(output).Decode(&info) }()
	select {
	case err := <-ready:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("нет готовности процесса")
	}
	return result
}

func TestWriteFailureIsReportedAndDoesNotAdvanceRevision(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	base := t.TempDir()
	testProfile(t, base, "io-error")
	process := startFailureServer(t, ctx, base)
	connection := storedArgs(base, base, "io-error")
	initial := commandSnapshot(t, ctx, connection, "session", "create", "--session", "main", "--name", "До отказа", "--request-id", "create")
	if err := os.WriteFile(filepath.Join(base, "fail-save"), []byte("fail"), 0600); err != nil {
		t.Fatal(err)
	}
	rename := []string{"session", "rename", "--session", "main", "--name", "После отказа", "--revision", "1", "--request-id", "rename"}
	cmd := exec.CommandContext(ctx, binary, append(rename, connection...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err == nil || stdout.Len() != 0 || !strings.Contains(stderr.String(), "storage_failure") {
		t.Fatalf("ошибка сохранения не сообщена: %v, %s", err, &stderr)
	}
	if got := commandSnapshot(t, ctx, connection, "session", "get", "--session", "main"); got != initial {
		t.Fatal("неподтверждённое изменение опубликовано")
	}
	if err := os.Remove(filepath.Join(base, "fail-save")); err != nil {
		t.Fatal(err)
	}
	if got := commandSnapshot(t, ctx, connection, rename...); got.Revision != 2 {
		t.Fatal("неуспешная команда зарезервировала request_id")
	}
	cli(t, ctx, 0, append([]string{"shutdown"}, connection...)...)
	select {
	case err := <-process:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("процесс не завершился")
	}
}
