package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/discovery"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/host"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/storage"
)

// Барьер существует только в тестовом процессе. Рабочий сервер не имеет таких флагов.
type recoveryJournal struct {
	*storage.Journal
	base string
}

func recoveryPhase(change application.Change) string {
	if change.Calculate == nil {
		return ""
	}
	switch change.Calculate.RequestID {
	case "before-commit":
		return "before"
	case "after-commit":
		return "after"
	default:
		return ""
	}
}

func (j recoveryJournal) Append(ctx context.Context, change application.Change) error {
	phase := recoveryPhase(change)
	if phase == "before" {
		if err := j.pause(ctx, change, phase); err != nil {
			return err
		}
	}
	if err := j.Journal.Append(ctx, change); err != nil {
		return err
	}
	if phase == "after" {
		return j.pauseCommitted(ctx, change)
	}
	return nil
}

func (j recoveryJournal) pauseCommitted(ctx context.Context, change application.Change) error {
	err := j.pause(ctx, change, "after")
	// Отмена после Sync прерывает доставку ответа, но не отменяет принятую команду.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	return err
}

func gatePath(base, phase, suffix string) string {
	return filepath.Join(base, "gate-"+phase+"."+suffix)
}

func (j recoveryJournal) pause(ctx context.Context, change application.Change, phase string) error {
	data, err := json.Marshal(change.Result)
	if err != nil {
		return err
	}
	if err := os.WriteFile(gatePath(j.base, phase, "json"), data, 0600); err != nil {
		return err
	}
	// Готовность появляется после закрытия файла: родитель не прочитает неполный JSON.
	if err := os.WriteFile(gatePath(j.base, phase, "ready"), nil, 0600); err != nil {
		return err
	}
	return j.wait(ctx, phase)
}

func (j recoveryJournal) wait(ctx context.Context, phase string) error {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := os.Stat(gatePath(j.base, phase, "release")); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			if err := os.WriteFile(gatePath(j.base, phase, "canceled"), nil, 0600); err != nil {
				return err
			}
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func TestRecoveryServerHelper(t *testing.T) {
	if os.Getenv("POB2_RECOVERY_HELPER") != "1" {
		t.Skip("вспомогательный процесс с барьером записи")
	}
	base := os.Getenv("POB2_HELPER_BASE")
	journal, err := storage.Open(t.Context(), base, "recovery")
	if err != nil {
		t.Fatal(err)
	}
	defer journal.Close()
	service, err := application.OpenService(t.Context(), recoveryJournal{Journal: journal, base: base})
	if err != nil {
		t.Fatal(err)
	}
	profile, err := discovery.Open(base, "recovery")
	if err != nil {
		t.Fatal(err)
	}
	server, err := host.Start(profile, "127.0.0.1:0", "", service, lifecycle.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer closeRecoveryServer(server)
	if err := json.NewEncoder(os.Stdout).Encode(server.Ready()); err != nil {
		t.Fatal(err)
	}
	<-server.Done()
	if err := server.Err(); err != nil {
		t.Fatal(err)
	}
}

func closeRecoveryServer(server *host.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = server.Close(ctx)
}

func newGatedRecoveryScenario(t *testing.T) *recoveryScenario {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)
	s := &recoveryScenario{t: t, ctx: ctx, dataDir: t.TempDir(), profile: "recovery"}
	s.process = startRecoveryProcess(t, ctx, s.dataDir)
	s.connect()
	return s
}

func startRecoveryProcess(t *testing.T, ctx context.Context, base string) *process {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(ctx)
	command := exec.CommandContext(ctx, executable, "-test.run=^TestRecoveryServerHelper$")
	command.Env = append(os.Environ(), "POB2_RECOVERY_HELPER=1", "POB2_HELPER_BASE="+base)
	command.Stderr = os.Stderr
	output, err := command.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	p := &process{cmd: command, done: make(chan struct{})}
	go func() { p.err = command.Wait(); close(p.done) }()
	t.Cleanup(func() { cancel(); <-p.done })
	ready := make(chan error, 1)
	go func() { ready <- json.NewDecoder(output).Decode(&p.ready) }()
	select {
	case err := <-ready:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("вспомогательный процесс не запустился")
	}
	return p
}

func (s *recoveryScenario) waitGate(phase, suffix string) {
	s.t.Helper()
	await(s.t, s.ctx, "барьер "+phase+" "+suffix, func() bool {
		_, err := os.Stat(gatePath(s.dataDir, phase, suffix))
		return err == nil
	})
}

func (s *recoveryScenario) gateResult(phase string) domain.Snapshot {
	s.t.Helper()
	s.waitGate(phase, "ready")
	data, err := os.ReadFile(gatePath(s.dataDir, phase, "json"))
	if err != nil {
		s.t.Fatal(err)
	}
	return decodeCatalogJSON[domain.Snapshot](s.t, data)
}
