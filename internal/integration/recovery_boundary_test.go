package integration_test

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/storage"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestThreeSpellCancellationAtCommitBoundary(t *testing.T) {
	for _, spell := range recoverySpells() {
		for _, adapter := range []string{"grpc", "mcp", "cli"} {
			for _, phase := range []string{"before", "after"} {
				t.Run(spell.id+"/"+adapter+"/"+phase, func(t *testing.T) { checkCanceledRecovery(t, spell, adapter, phase) })
			}
		}
	}
}

func checkCanceledRecovery(t *testing.T, spell recoverySpell, adapter, phase string) {
	t.Helper()
	s := newGatedRecoveryScenario(t)
	saved := s.importSpell(spell)
	command := application.CalculateBuild{SessionID: spell.id, ExpectedRevision: 2, RequestID: phase + "-commit"}
	cancel, answer := s.beginCalculation(command, adapter)
	pending := s.gateResult(phase)
	checkRecoveredCalculation(t, pending, spell, 1)
	s.assertNoEvents(spell.id, 2)
	cancel()
	s.expectInterrupted(answer)
	s.waitGate(phase, "canceled")
	s.waitIdle()
	want := saved.imported
	if phase == "after" {
		want = pending
	}
	if s.snapshot(spell.id) != want {
		t.Fatal("отмена нарушила границу сохранения")
	}
	if phase == "before" {
		s.assertNoEvents(spell.id, 2)
	}
	if err := os.WriteFile(gatePath(s.dataDir, phase, "release"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	result, err := s.client.Calculate(s.ctx, command)
	if err != nil || result != pending {
		t.Fatal("отменённый request_id не допускает безопасный повтор")
	}
	s.expectCalculationEvent(result, command.RequestID)
	s.kill()
	s.start()
	if s.snapshot(spell.id) != pending {
		t.Fatal("перезапуск потерял повторённый расчёт")
	}
	s.stop()
	s.verifyRecords(3)
}

func TestThreeSpellCrashAtCommitBoundary(t *testing.T) {
	for _, spell := range recoverySpells() {
		for _, phase := range []string{"before", "after"} {
			t.Run(spell.id+"/"+phase, func(t *testing.T) { checkCommitBoundary(t, spell, phase) })
		}
	}
}

func checkCommitBoundary(t *testing.T, spell recoverySpell, phase string) {
	t.Helper()
	s := newGatedRecoveryScenario(t)
	saved := s.importSpell(spell)
	command := application.CalculateBuild{SessionID: spell.id, ExpectedRevision: 2, RequestID: phase + "-commit"}
	cancel, answer := s.beginCalculation(command, "grpc")
	defer cancel()
	pending := s.gateResult(phase)
	s.assertNoEvents(spell.id, 2)
	s.kill()
	s.expectInterrupted(answer)
	s.start()
	want := saved.imported
	if phase == "after" {
		want = pending
	}
	if s.snapshot(spell.id) != want {
		t.Fatal("восстановление не соответствует границе commit")
	}
	result, err := s.client.Calculate(s.ctx, command)
	if err != nil || result != pending {
		t.Fatalf("потерян либо повторно применён результат запроса: %v,\nполучен %s\nожидался %s", err, result.CalculationJSON, pending.CalculationJSON)
	}
	checkRecoveredCalculation(t, result, spell, 1)
	s.expectCalculationEvent(result, command.RequestID)
	s.stop()
	s.verifyRecords(3)
}

func (s *recoveryScenario) beginCalculation(command application.CalculateBuild, adapter string) (context.CancelFunc, <-chan error) {
	s.t.Helper()
	ctx, cancel := context.WithCancel(s.ctx)
	s.t.Cleanup(cancel)
	answer := make(chan error, 1)
	switch adapter {
	case "grpc":
		go func() { _, err := s.client.Calculate(ctx, command); answer <- err }()
	case "mcp":
		client := s.attach(command.SessionID)
		go func() {
			_, err := client.CallTool(ctx, &mcp.CallToolParams{Name: "calculate_build", Arguments: map[string]any{"expected_revision": command.ExpectedRevision, "request_id": command.RequestID}})
			answer <- err
		}()
	case "cli":
		args := append([]string{"calculate", "--session", command.SessionID, "--revision", strconv.FormatUint(command.ExpectedRevision, 10), "--request-id", command.RequestID}, s.connection...)
		process := exec.CommandContext(ctx, binary, args...)
		go func() { answer <- process.Run() }()
	default:
		s.t.Fatalf("неизвестный адаптер %s", adapter)
	}
	return cancel, answer
}

func (s *recoveryScenario) expectInterrupted(answer <-chan error) {
	s.t.Helper()
	select {
	case err := <-answer:
		if err == nil {
			s.t.Fatal("прерванная команда вернула успешный ответ")
		}
	case <-s.ctx.Done():
		s.t.Fatal("клиент не завершил прерванную команду")
	}
}

func (s *recoveryScenario) waitIdle() {
	s.t.Helper()
	await(s.t, s.ctx, "завершение отменённой задачи", func() bool {
		info, err := grpcapi.Inspect(s.ctx, s.process.ready.Address, s.process.ready.Token)
		return err == nil && info.GetActiveTasks() == 0
	})
}

func (s *recoveryScenario) assertNoEvents(id string, revision uint64) {
	s.t.Helper()
	ctx, cancel := context.WithTimeout(s.ctx, 100*time.Millisecond)
	defer cancel()
	stream, err := s.client.Watch(ctx, id, revision)
	if err != nil {
		s.t.Fatal(err)
	}
	_, err = stream.Recv()
	if status.Code(err) != codes.DeadlineExceeded {
		s.t.Fatalf("неподтверждённое изменение создало событие: %v", err)
	}
}

func (s *recoveryScenario) expectCalculationEvent(want domain.Snapshot, request string) {
	s.t.Helper()
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()
	events, err := s.client.Watch(ctx, want.SessionID, 2)
	if err != nil {
		s.t.Fatal(err)
	}
	s.expectEvent(events, want, request, "calculated", "calculation")
	s.assertNoEvents(want.SessionID, 3)
}

func (s *recoveryScenario) verifyRecords(count uint64) {
	s.t.Helper()
	path, err := storage.Path(s.dataDir, s.profile)
	if err != nil {
		s.t.Fatal(err)
	}
	verified, err := storage.Verify(s.ctx, path)
	if err != nil {
		s.t.Fatal(err)
	}
	if verified.Schema != "2" || verified.Records != count {
		s.t.Fatalf("неверный журнал: %+v", verified)
	}
}
