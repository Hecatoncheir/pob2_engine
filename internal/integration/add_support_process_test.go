package integration_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAddSupportAcrossProcessesAndRestart(t *testing.T) {
	s := newGroupScenario(t)
	before := s.calculate("before-support", "Fireball", 1, 1, 8.9166666666667)
	edit := &build.Edit{AddSupport: &build.AddSupportEdit{SupportID: "SupportRapidCastingPlayer"}}
	preview, err := s.client.PreviewSupport(s.ctx, application.PreviewSupport{
		SessionID: "main", ExpectedRevision: before.Revision, SupportID: edit.AddSupport.SupportID,
	})
	if err != nil || !preview.Compatible {
		t.Fatalf("предпросмотр: %+v %v", preview, err)
	}
	if current := supportSnapshot(t, s); !reflect.DeepEqual(current, before) {
		t.Fatal("предпросмотр изменил сохранённый снимок")
	}
	command := application.ApplyBuild{SessionID: "main", RequestID: "add-rapid", ExpectedRevision: before.Revision, Edit: edit}
	added := addSupportCLI(t, s, command)
	if added.Build.InputSHA256 != preview.Comparison.After.InputSHA256 || added.Build.Calculation != nil {
		t.Fatal("применение отличается от предпросмотра или сохранило старый расчёт")
	}
	s.expectEvent(added.Revision, "build.model.groups")
	checkAddSupportConflicts(t, s, command, added)
	updated := addSupportMCP(t, s, added.Revision)
	s.expectEvent(added.Revision+1, "build.model.groups")
	checkBuildEditRestart(t, s, command, added, updated)
}

func addSupportCLI(t *testing.T, s *groupScenario, command application.ApplyBuild) application.SnapshotView {
	t.Helper()
	file := filepath.Join(t.TempDir(), "support.json")
	data, err := json.Marshal(command.Edit)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"build", "edit", "--session", command.SessionID, "--request-id", command.RequestID,
		"--revision", fmt.Sprint(command.ExpectedRevision), "--file", file}
	return buildView(t, cli(t, s.ctx, 0, append(args, s.connection...)...))
}

func supportSnapshot(t *testing.T, s *groupScenario) application.SnapshotView {
	t.Helper()
	snapshot, err := s.client.Snapshot(s.ctx, "main")
	if err != nil {
		t.Fatal(err)
	}
	view, err := application.ViewSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return view
}

func checkAddSupportConflicts(t *testing.T, s *groupScenario, command application.ApplyBuild, added application.SnapshotView) {
	t.Helper()
	checkSupportRetry(t, s, command, added)
	command.RequestID = "stale-support"
	if _, err := s.client.Apply(s.ctx, command); status.Code(err) != codes.Aborted {
		t.Fatalf("устаревшая ревизия: %v", err)
	}
	command.ExpectedRevision = added.Revision
	command.Edit = &build.Edit{AddSupport: &build.AddSupportEdit{SupportID: "FireballPlayer"}}
	if _, err := s.client.Apply(s.ctx, command); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("активный камень принят как поддержка: %v", err)
	}
}

func addSupportMCP(t *testing.T, s *groupScenario, revision uint64) application.SnapshotView {
	t.Helper()
	result := call(t, s.ctx, s.mcp, "edit_build", map[string]any{
		"expected_revision": revision, "request_id": "add-controlled",
		"edit": map[string]any{"add_support": map[string]any{"support_id": "Controlled Destruction"}},
	}, false)
	updated := structuredView(t, result.StructuredContent)
	if updated.Revision != revision+1 || len(updated.Build.Model.Groups[0].Gems) != 3 {
		t.Fatal("MCP не добавил одну поддержку")
	}
	return updated
}

func checkSupportRetry(t *testing.T, s *groupScenario, command application.ApplyBuild, want application.SnapshotView) {
	t.Helper()
	snapshot, err := s.client.Apply(s.ctx, command)
	if err != nil {
		t.Fatal(err)
	}
	view, err := application.ViewSnapshot(snapshot)
	if err != nil || !reflect.DeepEqual(view, want) {
		t.Fatalf("повтор gRPC вернул другой ответ: %v", err)
	}
}

func checkBuildEditRestart(t *testing.T, s *groupScenario, command application.ApplyBuild, added, updated application.SnapshotView) {
	t.Helper()
	_ = s.mcp.Close()
	_ = s.client.Close()
	if err := s.process.cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	<-s.process.done
	s.process = startServer(t, s.ctx, "--profile", "groups", "--runtime-dir", t.TempDir(), "--data-dir", s.dataDir)
	s.connect()
	checkSupportRetry(t, s, command, added)
	if current := supportSnapshot(t, s); !reflect.DeepEqual(current, updated) {
		t.Fatal("перезапуск или повтор изменил сохранённую сборку")
	}
}
