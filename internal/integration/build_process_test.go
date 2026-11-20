package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
)

func buildView(t *testing.T, data []byte) application.SnapshotView {
	t.Helper()
	var value application.SnapshotView
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func structuredView(t *testing.T, value any) application.SnapshotView {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return buildView(t, data)
}

func TestFullBuildThroughCLIAndMCPRecoversAfterRestart(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	base := t.TempDir()
	profile := "full-build"
	p := startServer(t, ctx, "--profile", profile, "--runtime-dir", base, "--data-dir", base)
	connection := []string{"--address", p.ready.Address, "--token", p.ready.Token}
	cli(t, ctx, 0, append([]string{"session", "create", "--session", "main", "--name", "Сборка", "--request-id", "create"}, connection...)...)
	grpc, err := grpcapi.Dial(ctx, p.ready.Address, p.ready.Token)
	if err != nil {
		t.Fatal(err)
	}
	stream, err := grpc.Watch(ctx, "main", 1)
	if err != nil {
		t.Fatal(err)
	}
	imported := buildView(t, cli(t, ctx, 0, append([]string{"build", "import", "--session", "main", "--revision", "1", "--request-id", "import", "--file", "../../testdata/reference/fireball.xml"}, connection...)...))
	if imported.Build == nil || imported.Build.Model.Class != "Sorceress" {
		t.Fatal("CLI не импортировал полную сборку")
	}
	first, err := stream.Recv()
	if err != nil || first.GetSnapshot().GetRevision() != 2 {
		t.Fatalf("импорт не доставлен: %v %v", first, err)
	}
	m := connectMCP(t, ctx, append([]string{"--attach", "main"}, connection...)...)
	edit := map[string]any{"edit": map[string]any{"config": map[string]any{"name": "customMods", "value": map[string]any{"type": "string", "value": "100% increased Spell Damage"}}}, "expected_revision": 2, "request_id": "damage"}
	changed := structuredView(t, call(t, ctx, m, "edit_build", edit, false).StructuredContent)
	if changed.Revision != 3 || changed.Build.Calculation != nil {
		t.Fatal("MCP не изменил вход")
	}
	second, err := stream.Recv()
	if err != nil || !slices.Contains(second.GetChangedFields(), "build.model.config") {
		t.Fatalf("нет точного поля изменения: %v %v", second, err)
	}
	calculate := []string{"calculate", "--session", "main", "--revision", "3", "--request-id", "calculate"}
	saved := buildView(t, cli(t, ctx, 0, append(calculate, connection...)...))
	if saved.Build.Calculation.Status != "complete" || math.Abs(saved.Build.Calculation.Stats["TotalDPS"]-17.8333333333333) > 1e-9 {
		t.Fatalf("неверный расчёт: %+v", saved.Build.Calculation)
	}
	third, err := stream.Recv()
	if err != nil || third.GetKind() != "calculated" || third.GetSnapshot().GetBuild().GetCalculation() == nil {
		t.Fatal("результат не передан по подписке")
	}
	fromMCP := structuredView(t, call(t, ctx, m, "get_snapshot", map[string]any{}, false).StructuredContent)
	fromGRPC, err := grpc.Snapshot(ctx, "main")
	if err != nil {
		t.Fatal(err)
	}
	grpcView, err := application.ViewSnapshot(fromGRPC)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(fromMCP, saved) || !reflect.DeepEqual(grpcView, saved) {
		t.Fatal("CLI, MCP и gRPC видят разные сборки")
	}
	compare := call(t, ctx, m, "compare_build", map[string]any{"expected_revision": 4, "edit": map[string]any{"config": map[string]any{"name": "customMods", "value": map[string]any{"type": "string", "value": "200% increased Spell Damage"}}}}, false)
	data, err := json.Marshal(compare.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var compared application.BuildComparison
	if err := json.Unmarshal(data, &compared); err != nil {
		t.Fatal(err)
	}
	if compared.Comparison.After.Stats["TotalDPS"] <= compared.Comparison.Before.Stats["TotalDPS"] {
		t.Fatal("вариант не сравнивается")
	}
	if actual := structuredView(t, call(t, ctx, m, "get_snapshot", map[string]any{}, false).StructuredContent); !reflect.DeepEqual(actual, saved) {
		t.Fatal("сравнение изменило сессию")
	}
	call(t, ctx, m, "edit_build", edit, false)
	if actual := structuredView(t, call(t, ctx, m, "get_snapshot", map[string]any{}, false).StructuredContent); actual.Revision != 4 {
		t.Fatal("повтор создаёт новую правку")
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	grpc.Close()
	cli(t, ctx, 0, append([]string{"shutdown"}, connection...)...)
	select {
	case <-p.done:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	backup := filepath.Join(t.TempDir(), "saved.db")
	cli(t, ctx, 0, "storage", "backup", "--profile", profile, "--data-dir", base, "--file", backup)
	cli(t, ctx, 0, "storage", "verify", "--file", backup)
	copyBase := t.TempDir()
	cli(t, ctx, 0, "storage", "restore", "--profile", "restored", "--data-dir", copyBase, "--file", backup)
	restarted := startServer(t, ctx, "--profile", "restored", "--runtime-dir", t.TempDir(), "--data-dir", copyBase)
	connection = []string{"--address", restarted.ready.Address, "--token", restarted.ready.Token}
	replayed := buildView(t, cli(t, ctx, 0, append(calculate, connection...)...))
	if !reflect.DeepEqual(saved, replayed) {
		t.Fatal("перезапуск потерял исходный результат")
	}
	exported := cli(t, ctx, 0, append([]string{"build", "export", "--session", "main", "--raw"}, connection...)...)
	if string(exported) != saved.Build.XML {
		t.Fatal("экспорт отличается от сохраняемого XML")
	}
	output := filepath.Join(t.TempDir(), "saved.xml")
	cli(t, ctx, 0, append([]string{"build", "export", "--session", "main", "--output", output}, connection...)...)
	data, err = os.ReadFile(output)
	if err != nil || string(data) != saved.Build.XML {
		t.Fatalf("файл экспорта не совпадает: %v", err)
	}
	cli(t, ctx, 1, append([]string{"build", "export", "--session", "main", "--output", output}, connection...)...)
	if err := restarted.cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	<-restarted.done
	recovered := startServer(t, ctx, "--profile", "restored", "--runtime-dir", t.TempDir(), "--data-dir", copyBase)
	connection = []string{"--address", recovered.ready.Address, "--token", recovered.ready.Token}
	afterCrash := buildView(t, cli(t, ctx, 0, append(calculate, connection...)...))
	if !reflect.DeepEqual(saved, afterCrash) {
		t.Fatal("аварийный перезапуск потерял сборку или результат запроса")
	}
}

func TestCleanStandaloneCalculationAndStatusExitCodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	for _, sample := range []struct {
		name   string
		code   int
		status string
	}{
		{"fireball.xml", 0, "complete"}, {"cases/gem-requirement-warning.xml", 4, "partial"}, {"cases/unparsed-custom-modifier.xml", 3, "unsupported"},
	} {
		t.Run(sample.status, func(t *testing.T) {
			root := t.TempDir()
			data, err := os.ReadFile(filepath.Join("../../testdata/reference", sample.name))
			if err != nil {
				t.Fatal(err)
			}
			input := filepath.Join(root, "build.xml")
			if err := os.WriteFile(input, data, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(ctx, binary, "calculate", "--file", input)
			cmd.Dir = root
			for _, entry := range os.Environ() {
				name, _, _ := strings.Cut(entry, "=")
				if name != "PATH" && name != "HOME" && name != "APPDATA" && name != "LOCALAPPDATA" && name != "USERPROFILE" && !strings.HasPrefix(name, "POB_") && !strings.HasPrefix(name, "POB2_") && !strings.HasPrefix(name, "PYTHON") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, "PATH=", "HOME="+root, "USERPROFILE="+root, "APPDATA="+root, "LOCALAPPDATA="+root, "POB_DIR="+filepath.Join(root, "missing-pob"))
			var output, diagnostics bytes.Buffer
			cmd.Stdout = &output
			cmd.Stderr = &diagnostics
			_ = cmd.Run()
			if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != sample.code {
				t.Fatalf("выход: %v %s", cmd.ProcessState, &diagnostics)
			}
			var result calculation.Result
			if err := json.Unmarshal(output.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Status != sample.status || result.DataSHA256 == "" {
				t.Fatalf("неверный результат: %+v", result)
			}
			if sample.status == "unsupported" && result.Stats != nil {
				t.Fatal("неподдержанный ввод превратился в числа")
			}
			if sample.code != 0 && diagnostics.Len() == 0 {
				t.Fatal("нет диагностики в stderr")
			}
		})
	}
}
