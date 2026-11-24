package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"slices"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func decodeCatalogJSON[T any](t *testing.T, data []byte) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func structuredCatalogJSON[T any](t *testing.T, value any) T {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return decodeCatalogJSON[T](t, data)
}

func TestCatalogAndSupportPreviewAcrossProcesses(t *testing.T) {
	scenario := newGroupScenario(t)
	checkCatalogTools(t, scenario)
	for _, query := range []application.CatalogQuery{
		{Kind: "gems"}, {Kind: "classes"}, {Kind: "items"}, {Kind: "passives"},
		{Kind: "gems", Query: " rapidcastingsupport ", Limit: 1}, {Kind: "gems", Offset: 2, Limit: 2},
		{Kind: "items", Query: "no-match"}, {Kind: "passives", Offset: 10000},
	} {
		checkCatalogAdapters(t, scenario, query)
	}
	checkSupportGroups(t, scenario)
	checkPreviewDiagnosticCases(t, scenario)
	checkCatalogAdapterErrors(t, scenario)
	_, err := scenario.client.Rename(scenario.ctx, application.RenameBuild{SessionID: "main", ExpectedRevision: scenario.revision, RequestID: "after-preview", BuildName: "После предпросмотра"})
	if err != nil {
		t.Fatal(err)
	}
	scenario.expectEvent(scenario.revision+1, "build_name")
}

func checkCatalogTools(t *testing.T, s *groupScenario) {
	t.Helper()
	tools, err := s.mcp.ListTools(s.ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"get_catalog", "preview_support"} {
		index := slices.IndexFunc(tools.Tools, func(tool *mcp.Tool) bool { return tool.Name == name })
		if index < 0 {
			t.Fatalf("нет инструмента %s", name)
		}
		if tools.Tools[index].Annotations == nil || !tools.Tools[index].Annotations.ReadOnlyHint {
			t.Fatal("не объявлено чтение без правок")
		}
	}
}

func checkCatalogAdapters(t *testing.T, s *groupScenario, query application.CatalogQuery) {
	t.Helper()
	grpc, err := s.client.Catalog(s.ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	data := cli(t, s.ctx, 0, append([]string{"catalog", "--kind", query.Kind, "--query", query.Query, "--offset", fmt.Sprint(query.Offset), "--limit", fmt.Sprint(query.Limit)}, s.connection...)...)
	fromCLI := decodeCatalogJSON[application.CatalogPage](t, data)
	result := call(t, s.ctx, s.mcp, "get_catalog", map[string]any{"kind": query.Kind, "query": query.Query, "offset": query.Offset, "limit": query.Limit}, false)
	fromMCP := structuredCatalogJSON[application.CatalogPage](t, result.StructuredContent)
	if !reflect.DeepEqual(grpc, fromCLI) || !reflect.DeepEqual(grpc, fromMCP) {
		t.Fatalf("разные каталоги для %+v", query)
	}
}

func checkSupportGroups(t *testing.T, s *groupScenario) {
	t.Helper()
	spells := []string{"Fireball", "Frostbolt", "Spark"}
	for i, spell := range spells {
		if i > 0 {
			s.setEnabled(i, spells[i-1], false)
			s.setEnabled(i+1, spell, true)
			s.selectSkills("preview-"+spell, build.SelectionEdit{MainGroup: i + 1}, "build.model.main_group")
		}
		checkPreviewAdapters(t, s, "Rapid Casting I", "complete")
		checkPreviewAdapters(t, s, "Controlled Destruction", "complete")
	}
	installPreviewSupport(t, s)
	checkPreviewAdapters(t, s, "Controlled Destruction", "complete")
	preview := checkPreviewAdapters(t, s, "Rapid Casting I", "partial")
	if preview.Compatible || !slices.ContainsFunc(preview.Comparison.After.Issues, func(issue build.Issue) bool {
		return issue.Code == "pob_support_superseded"
	}) {
		t.Fatal("PoB не сообщил о вытесненной поддержке")
	}
}

func installPreviewSupport(t *testing.T, s *groupScenario) {
	t.Helper()
	edit := build.Edit{Gems: &build.GemsEdit{GroupID: 3, Enabled: true, Gems: []build.GemInput{
		{ID: "SparkPlayer", Level: 1, Enabled: true}, {ID: "Rapid Casting I", Level: 1, Enabled: true},
	}}}
	_, err := s.client.Apply(s.ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: "rapid-installed", Edit: &edit})
	if err != nil {
		t.Fatal(err)
	}
	s.revision++
	s.expectEvent(s.revision, "build.model.groups")
}

func checkPreviewAdapters(t *testing.T, s *groupScenario, support, state string) application.SupportPreview {
	t.Helper()
	before, err := s.client.Snapshot(s.ctx, "main")
	if err != nil {
		t.Fatal(err)
	}
	grpc, err := s.client.PreviewSupport(s.ctx, application.PreviewSupport{SessionID: "main", ExpectedRevision: s.revision, SupportID: support})
	if err != nil {
		t.Fatal(err)
	}
	result := call(t, s.ctx, s.mcp, "preview_support", map[string]any{"support_id": support, "expected_revision": s.revision}, false)
	fromMCP := structuredCatalogJSON[application.SupportPreview](t, result.StructuredContent)
	if grpc.Status != state {
		t.Fatalf("предпросмотр группы %d, %s: статус %s вместо %s; диагностика: %v / %v", grpc.GroupID, support, grpc.Status, state, grpc.Comparison.Before.Issues, grpc.Comparison.After.Issues)
	}
	fromCLI := previewCLI(t, s, support, state)
	if !reflect.DeepEqual(grpc, fromCLI) || !reflect.DeepEqual(grpc, fromMCP) {
		t.Fatalf("разные результаты предпросмотра %s: %+v", support, grpc)
	}
	after, err := s.client.Snapshot(s.ctx, "main")
	if err != nil || before != after {
		t.Fatal("предпросмотр изменил сессию")
	}
	return grpc
}

func previewCLI(t *testing.T, s *groupScenario, support, state string) application.SupportPreview {
	t.Helper()
	args := append([]string{"preview-support", "--session", "main", "--revision", fmt.Sprint(s.revision), "--support", support}, s.connection...)
	command := exec.CommandContext(s.ctx, binary, args...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	exit := 0
	if state == "unsupported" {
		exit = 3
	}
	if state == "partial" {
		exit = 4
	}
	if command.ProcessState == nil || command.ProcessState.ExitCode() != exit {
		t.Fatalf("CLI: %v %s", err, &stderr)
	}
	return decodeCatalogJSON[application.SupportPreview](t, stdout.Bytes())
}

func checkCatalogAdapterErrors(t *testing.T, s *groupScenario) {
	t.Helper()
	_, err := s.client.Catalog(s.ctx, application.CatalogQuery{Kind: "gems", Limit: 101})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatal("gRPC допустил слишком большую страницу")
	}
	call(t, s.ctx, s.mcp, "get_catalog", map[string]any{"kind": "gems", "limit": 101}, true)
	cli(t, s.ctx, 1, append([]string{"catalog", "--kind", "gems", "--limit", "101"}, s.connection...)...)
	_, err = s.client.PreviewSupport(s.ctx, application.PreviewSupport{SessionID: "main", ExpectedRevision: 1, SupportID: "Rapid Casting I"})
	if status.Code(err) != codes.Aborted {
		t.Fatal("gRPC не отклонил старую ревизию")
	}
	call(t, s.ctx, s.mcp, "preview_support", map[string]any{"support_id": "Rapid Casting I", "expected_revision": 1}, true)
	cli(t, s.ctx, 1, append([]string{"preview-support", "--session", "main", "--revision", "1", "--support", "Rapid Casting I"}, s.connection...)...)
}

func checkPreviewDiagnosticCases(t *testing.T, s *groupScenario) {
	t.Helper()
	for _, sample := range []struct{ file, state string }{
		{"gem-requirement-warning", "partial"}, {"unparsed-custom-modifier", "partial"},
	} {
		xml, err := os.ReadFile("../../testdata/reference/cases/" + sample.file + ".xml")
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.client.Apply(s.ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: s.revision, RequestID: sample.file, XML: string(xml)})
		if err != nil {
			t.Fatal(err)
		}
		s.revision++
		s.expectEvent(s.revision, "build.model.groups")
		checkPreviewAdapters(t, s, "Rapid Casting I", sample.state)
	}
}
