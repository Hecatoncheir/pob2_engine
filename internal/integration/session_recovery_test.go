package integration_test

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/storage"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type recoverySpell struct {
	id, name, file string
	dps            float64
}

func recoverySpells() []recoverySpell {
	return []recoverySpell{
		{"fireball", "Fireball", "../../testdata/reference/fireball.xml", 8.9166666666667},
		{"frostbolt", "Frostbolt", "../../testdata/reference/0.0.2/cases/frostbolt-minimal.xml", 9.1},
		{"spark", "Spark", "../../testdata/reference/0.0.2/cases/spark-minimal.xml", 8.5642857142857},
	}
}

type recoverySession struct {
	spell                                recoverySpell
	xml                                  string
	imported, edited, calculated, latest domain.Snapshot
}

type recoveryScenario struct {
	t                *testing.T
	ctx              context.Context
	process          *process
	client           *grpcapi.Client
	connection       []string
	dataDir, profile string
}

func newRecoveryScenario(t *testing.T) *recoveryScenario {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	t.Cleanup(cancel)
	s := &recoveryScenario{t: t, ctx: ctx, dataDir: t.TempDir(), profile: "recovery"}
	s.start()
	return s
}

func (s *recoveryScenario) start() {
	s.t.Helper()
	s.process = startServer(s.t, s.ctx, "--profile", s.profile, "--runtime-dir", s.t.TempDir(), "--data-dir", s.dataDir)
	s.connect()
}

func (s *recoveryScenario) connect() {
	s.t.Helper()
	s.connection = []string{"--address", s.process.ready.Address, "--token", s.process.ready.Token, "--timeout", "8s"}
	client, err := grpcapi.Dial(s.ctx, s.process.ready.Address, s.process.ready.Token)
	if err != nil {
		s.t.Fatal(err)
	}
	s.client = client
	s.t.Cleanup(func() { _ = client.Close() })
}

func (s *recoveryScenario) attach(id string) *mcp.ClientSession {
	s.t.Helper()
	return connectMCP(s.t, s.ctx, append([]string{"--attach", id}, s.connection...)...)
}

func spellDamageEdit() *build.Edit {
	return &build.Edit{Config: &build.ConfigEdit{Name: "customMods", Value: &build.ConfigValue{Type: "string", Value: "100% increased Spell Damage"}}}
}

func (s *recoveryScenario) snapshot(id string) domain.Snapshot {
	s.t.Helper()
	value, err := s.client.Snapshot(s.ctx, id)
	if err != nil {
		s.t.Fatal(err)
	}
	return value
}

func (s *recoveryScenario) importSpell(spell recoverySpell) recoverySession {
	s.t.Helper()
	xml, err := os.ReadFile(spell.file)
	if err != nil {
		s.t.Fatal(err)
	}
	_, err = s.client.Create(s.ctx, application.CreateSession{SessionID: spell.id, BuildName: spell.name, RequestID: "create-" + spell.id})
	if err != nil {
		s.t.Fatal(err)
	}
	cli(s.t, s.ctx, 0, append([]string{"build", "import", "--session", spell.id, "--revision", "1", "--request-id", "import", "--file", spell.file}, s.connection...)...)
	return recoverySession{spell: spell, xml: string(xml), imported: s.snapshot(spell.id)}
}

func (s *recoveryScenario) acceptSpell(spell recoverySpell) recoverySession {
	s.t.Helper()
	saved := s.importSpell(spell)
	events, err := s.client.Watch(s.ctx, spell.id, 1)
	if err != nil {
		s.t.Fatal(err)
	}
	s.expectEvent(events, saved.imported, "import", "build_changed", "build.model.groups")
	m := s.attach(spell.id)
	defer m.Close()
	call(s.t, s.ctx, m, "edit_build", map[string]any{"edit": spellDamageEdit(), "expected_revision": 2, "request_id": "damage"}, false)
	saved.edited = s.snapshot(spell.id)
	s.expectEvent(events, saved.edited, "damage", "build_changed", "build.model.config")
	saved.calculated, err = s.client.Calculate(s.ctx, application.CalculateBuild{SessionID: spell.id, ExpectedRevision: 3, RequestID: "calculate"})
	if err != nil {
		s.t.Fatal(err)
	}
	checkRecoveredCalculation(s.t, saved.calculated, spell, 2)
	s.expectEvent(events, saved.calculated, "calculate", "calculated", "calculation")
	s.assertAdapters(m, saved.calculated)
	s.rejectConflicts(m, saved)
	saved.latest, err = s.client.Rename(s.ctx, application.RenameBuild{SessionID: spell.id, ExpectedRevision: 4, RequestID: "rename", BuildName: "Сохранённая " + spell.name})
	if err != nil {
		s.t.Fatal(err)
	}
	s.expectEvent(events, saved.latest, "rename", "renamed", "build_name")
	return saved
}

func (s *recoveryScenario) expectEvent(events groupEvents, want domain.Snapshot, request, kind, field string) {
	s.t.Helper()
	event, err := events.Recv()
	if err != nil {
		s.t.Fatal(err)
	}
	if event.GetRequestId() != request || event.GetKind() != kind || !slices.Contains(event.GetChangedFields(), field) {
		s.t.Fatalf("неверное событие: %v", event)
	}
	s.checkEventSnapshot(event.GetSnapshot(), want)
	data, err := s.client.Snapshot(s.ctx, want.SessionID)
	if err != nil || data != want {
		s.t.Fatal("событие опережает подтверждённый снимок")
	}
}

func (s *recoveryScenario) checkEventSnapshot(got *pb.SessionSnapshot, want domain.Snapshot) {
	s.t.Helper()
	if got.GetSessionId() != want.SessionID || got.GetRevision() != want.Revision || got.GetBuildName() != want.BuildName {
		s.t.Fatal("событие передаёт другой снимок")
	}
	if got.GetBuild().GetXml() != want.BuildXML || got.GetBuild().GetInputSha256() != want.InputSHA256 {
		s.t.Fatal("событие передаёт другой вход")
	}
	view, err := application.ViewSnapshot(want)
	if err != nil {
		s.t.Fatal(err)
	}
	if view.Build.Calculation == nil {
		if got.GetBuild().GetCalculation() != nil {
			s.t.Fatal("событие содержит устаревший расчёт")
		}
		return
	}
	checkEventCalculation(s.t, got.GetBuild(), view.Build.Calculation, want.CalculatedFromRevision)
}

func checkRecoveredCalculation(t *testing.T, snapshot domain.Snapshot, spell recoverySpell, multiplier float64) {
	t.Helper()
	result, err := application.DecodeCalculation(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if result.Skill != spell.name || result.Status != "complete" || math.Abs(result.Stats["TotalDPS"]-spell.dps*multiplier) > 1e-9 {
		t.Fatalf("неверный расчёт %s: %+v", spell.name, result)
	}
	if result.InputSHA256 != snapshot.InputSHA256 || snapshot.CalculatedFromRevision != snapshot.BuildRevision {
		t.Fatal("потеряна привязка расчёта к входу")
	}
}

func checkEventCalculation(t *testing.T, got *pb.BuildDocument, want *calculation.Result, revision uint64) {
	t.Helper()
	result := got.GetCalculation()
	if result.GetStatus() != pb.CalculationStatus_CALCULATION_STATUS_COMPLETE || result.GetSkill() != want.Skill || result.GetMode() != want.Mode {
		t.Fatal("событие содержит другой расчёт")
	}
	if !reflect.DeepEqual(result.GetStats(), want.Stats) || result.GetInputSha256() != want.InputSHA256 || got.GetCalculatedFromRevision() != revision {
		t.Fatal("событие потеряло числа или привязку входа")
	}
	if result.GetDataId() != want.DataID || result.GetDataSha256() != want.DataSHA256 || result.GetPobCommit() != want.PoBCommit {
		t.Fatal("событие изменило происхождение расчёта")
	}
}

func (s *recoveryScenario) assertAdapters(m *mcp.ClientSession, want domain.Snapshot) {
	s.t.Helper()
	view, err := application.ViewSnapshot(want)
	if err != nil {
		s.t.Fatal(err)
	}
	fromCLI := buildView(s.t, cli(s.t, s.ctx, 0, append([]string{"session", "get", "--session", want.SessionID}, s.connection...)...))
	fromMCP := structuredView(s.t, call(s.t, s.ctx, m, "get_snapshot", map[string]any{}, false).StructuredContent)
	if !reflect.DeepEqual(view, fromCLI) || !reflect.DeepEqual(view, fromMCP) {
		s.t.Fatal("три адаптера видят разные сессии")
	}
}

func (s *recoveryScenario) rejectConflicts(m *mcp.ClientSession, saved recoverySession) {
	s.t.Helper()
	_, err := s.client.Apply(s.ctx, application.ApplyBuild{SessionID: saved.spell.id, ExpectedRevision: 2, RequestID: "stale", Edit: spellDamageEdit()})
	if status.Code(err) != codes.Aborted {
		s.t.Fatalf("не отклонена старая ревизия: %v", err)
	}
	_, err = s.client.Calculate(s.ctx, application.CalculateBuild{SessionID: saved.spell.id, ExpectedRevision: 4, RequestID: "damage"})
	if status.Code(err) != codes.FailedPrecondition {
		s.t.Fatalf("не отклонён занятый request_id: %v", err)
	}
	call(s.t, s.ctx, m, "edit_build", map[string]any{"edit": spellDamageEdit(), "expected_revision": 2, "request_id": "stale-mcp"}, true)
	cli(s.t, s.ctx, 1, append([]string{"calculate", "--session", saved.spell.id, "--revision", "2", "--request-id", "stale-cli"}, s.connection...)...)
	if s.snapshot(saved.spell.id) != saved.calculated {
		s.t.Fatal("конфликт изменил сессию")
	}
}

func TestThreeSpellSessionsSurviveCrashAndBackupRestore(t *testing.T) {
	s := newRecoveryScenario(t)
	var saved []recoverySession
	for _, spell := range recoverySpells() {
		saved = append(saved, s.acceptSpell(spell))
	}
	instance := s.process.ready.InstanceID
	s.kill()
	s.start()
	if s.process.ready.InstanceID == instance {
		t.Fatal("сервер не перезапускался")
	}
	s.verifySessions(saved)
	s.stop()
	s.restoreBackup(15)
	s.start()
	s.verifySessions(saved)
	s.stop()
	s.verifyRecords(15)
}

func (s *recoveryScenario) kill() {
	s.t.Helper()
	if err := s.process.cmd.Process.Kill(); err != nil {
		s.t.Fatal(err)
	}
	select {
	case <-s.process.done:
	case <-s.ctx.Done():
		s.t.Fatal(s.ctx.Err())
	}
	_ = s.client.Close()
}

func (s *recoveryScenario) stop() {
	s.t.Helper()
	_ = s.client.Close()
	cli(s.t, s.ctx, 0, append([]string{"shutdown"}, s.connection...)...)
	select {
	case <-s.process.done:
	case <-s.ctx.Done():
		s.t.Fatal(s.ctx.Err())
	}
	if s.process.err != nil {
		s.t.Fatal(s.process.err)
	}
}

func (s *recoveryScenario) restoreBackup(records uint64) {
	s.t.Helper()
	backup := filepath.Join(s.t.TempDir(), "sessions.db")
	cli(s.t, s.ctx, 0, "storage", "backup", "--profile", s.profile, "--data-dir", s.dataDir, "--file", backup)
	data := cli(s.t, s.ctx, 0, "storage", "verify", "--file", backup)
	verified := decodeCatalogJSON[storage.Verification](s.t, data)
	if verified.Schema != "2" || verified.Records != records {
		s.t.Fatalf("изменён формат или журнал: %+v", verified)
	}
	s.dataDir = s.t.TempDir()
	s.profile = "restored"
	cli(s.t, s.ctx, 0, "storage", "restore", "--profile", s.profile, "--data-dir", s.dataDir, "--file", backup)
}

func (s *recoveryScenario) verifySessions(saved []recoverySession) {
	s.t.Helper()
	for _, session := range saved {
		s.verifySession(session)
	}
}

func (s *recoveryScenario) verifySession(saved recoverySession) {
	s.t.Helper()
	id := saved.spell.id
	if s.snapshot(id) != saved.latest {
		s.t.Fatal("восстановление изменило снимок")
	}
	m := s.attach(id)
	defer m.Close()
	s.assertAdapters(m, saved.latest)
	replayed, err := s.client.Apply(s.ctx, application.ApplyBuild{SessionID: id, ExpectedRevision: 1, RequestID: "import", XML: saved.xml})
	if err != nil || replayed != saved.imported {
		s.t.Fatal("потерян первоначальный результат импорта")
	}
	s.replayEdit(m, saved)
	data := cli(s.t, s.ctx, 0, append([]string{"calculate", "--session", id, "--revision", "3", "--request-id", "calculate"}, s.connection...)...)
	view, err := application.ViewSnapshot(saved.calculated)
	if err != nil || !reflect.DeepEqual(buildView(s.t, data), view) {
		s.t.Fatal("потерян первоначальный результат расчёта")
	}
	if s.snapshot(id) != saved.latest {
		s.t.Fatal("повтор откатил текущий снимок")
	}
	s.verifyHistory(saved)
}

func (s *recoveryScenario) verifyHistory(saved recoverySession) {
	s.t.Helper()
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()
	events, err := s.client.Watch(ctx, saved.spell.id, 1)
	if err != nil {
		s.t.Fatal(err)
	}
	for _, want := range []domain.Snapshot{saved.imported, saved.edited, saved.calculated, saved.latest} {
		event, err := events.Recv()
		if err != nil {
			s.t.Fatal(err)
		}
		s.checkEventSnapshot(event.GetSnapshot(), want)
	}
	s.assertNoEvents(saved.spell.id, 5)
}

func (s *recoveryScenario) replayEdit(client *mcp.ClientSession, saved recoverySession) {
	s.t.Helper()
	result := call(s.t, s.ctx, client, "edit_build", map[string]any{"edit": spellDamageEdit(), "expected_revision": 2, "request_id": "damage"}, false)
	want, err := application.ViewSnapshot(saved.edited)
	if err != nil {
		s.t.Fatal(err)
	}
	if !reflect.DeepEqual(structuredView(s.t, result.StructuredContent), want) {
		s.t.Fatal("потерян первоначальный ответ правки")
	}
}
