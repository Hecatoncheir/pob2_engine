package integration_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type workshopReport struct {
	Revision uint64         `json:"revision"`
	Report   map[string]any `json:"report"`
}

func checkFirstReportAfterEdit(t *testing.T, s *groupScenario, before, applied application.SnapshotView) {
	t.Helper()
	prepared := reportJSON(t, s, "main", applied.Revision)
	other, err := s.client.Create(s.ctx, application.CreateSession{
		SessionID: "report-before-edit", BuildName: before.BuildName,
		RequestID: "report-before-edit", XML: before.Build.XML,
	})
	if err != nil {
		t.Fatal(err)
	}
	reportJSON(t, s, other.SessionID, other.Revision)
	if reloaded := reportJSON(t, s, "main", applied.Revision); !reflect.DeepEqual(prepared, reloaded) {
		t.Fatal("подготовленный отчёт отличается от отчёта после повторной загрузки XML")
	}
}

func reportJSON(t *testing.T, s *groupScenario, id string, revision uint64) map[string]any {
	t.Helper()
	view, err := s.client.BuildReport(s.ctx, id, revision)
	if err != nil || view.Revision != revision {
		t.Fatal("не получен отчёт нужной ревизии", err)
	}
	return decodeCatalogJSON[map[string]any](t, view.Report)
}

func TestLuaBuildReportAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenario(t, "tree-specs/warrior-witch.xml")
	before := supportSnapshot(t, s)
	view := compareReportTransports(t, s)
	checkWorkshopReport(t, view.Report)
	if _, err := s.client.BuildReport(s.ctx, "main", s.revision-1); status.Code(err) != codes.Aborted {
		t.Fatal("принята старая ревизия отчёта", err)
	}
	if !reflect.DeepEqual(supportSnapshot(t, s), before) {
		t.Fatal("чтение отчёта изменило состояние сборки")
	}
	checkReportAfterRename(t, s, before)
}

func checkReportAfterRename(t *testing.T, s *groupScenario, before application.SnapshotView) {
	t.Helper()
	const name = "Новое имя для прежнего XML"
	renamed, err := s.client.Rename(s.ctx, application.RenameBuild{
		SessionID: "main", BuildName: name, ExpectedRevision: s.revision, RequestID: "report-rename",
	})
	if err != nil || renamed.BuildXML != before.Build.XML {
		t.Fatal("переименование изменило XML или завершилось ошибкой", err)
	}
	view, err := s.client.BuildReport(s.ctx, "main", renamed.Revision)
	if err != nil || view.Revision != renamed.Revision {
		t.Fatal("не получен отчёт новой ревизии", err)
	}
	type reportName struct {
		Snapshot struct {
			BuildInfo struct{ Name string }
		}
	}
	report := decodeCatalogJSON[reportName](t, view.Report)
	if got := report.Snapshot.BuildInfo.Name; got != name {
		t.Fatalf("в отчёте имя %q, ожидалось %q", got, name)
	}
}

func compareReportTransports(t *testing.T, s *groupScenario) workshopReport {
	t.Helper()
	reply, err := s.client.BuildReport(s.ctx, "main", s.revision)
	if err != nil {
		t.Fatal(err)
	}
	view := workshopReport{Revision: reply.Revision, Report: decodeCatalogJSON[map[string]any](t, reply.Report)}
	args := []string{"build-report", "--session", "main", "--revision", fmt.Sprint(s.revision)}
	cliReport := decodeCatalogJSON[workshopReport](t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	answer := call(t, s.ctx, s.mcp, "get_build_report", map[string]any{"expected_revision": s.revision}, false)
	mcpReport := structuredCatalogJSON[workshopReport](t, answer.StructuredContent)
	if !reflect.DeepEqual(view, cliReport) || !reflect.DeepEqual(view, mcpReport) {
		t.Fatal("подробные отчёты CLI, MCP и gRPC различаются")
	}
	return view
}

func checkWorkshopReport(t *testing.T, report map[string]any) {
	t.Helper()
	if report["source"] != "original-lua" || !strings.Contains(fmt.Sprint(report["notes"]), "Обе спецификации") {
		t.Fatal("отчёт не содержит заметку контрольной сборки")
	}
	if !strings.Contains(fmt.Sprint(report["groups"]), "Rolling Slam") || !strings.Contains(fmt.Sprint(report["items"]), "Wooden Club") {
		t.Fatal("из отчёта пропали камни или экипировка")
	}
	if !strings.Contains(fmt.Sprint(report["config"]), "enemyLevel:40") || !strings.Contains(fmt.Sprint(report["sets"]), "Ведьма") {
		t.Fatal("из отчёта пропали условия или спецификация дерева")
	}
	definitions, ok := report["conditionDefinitions"].([]any)
	if !ok || len(definitions) < 20 || !strings.Contains(fmt.Sprint(definitions), "id:enemyLevel") {
		t.Fatal("в отчёте нет каталога условий Lua")
	}
}
