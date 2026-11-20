package integration_test

import (
	"context"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
)

type spellSample struct {
	file, skill string
	dps         float64
}

func TestNewSpellsAgreeAcrossCLIAndMCPAndGRPC(t *testing.T) {
	for _, sample := range []spellSample{
		{"frostbolt", "Frostbolt", 9.1}, {"spark", "Spark", 8.5642857142857},
	} {
		t.Run(sample.skill, func(t *testing.T) {
			checkSpellAcrossAdapters(t, sample)
		})
	}
}

func checkSpellAcrossAdapters(t *testing.T, sample spellSample) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	path := "../../testdata/reference/0.0.2/cases/" + sample.file + "-minimal.xml"
	xml, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p := startServer(t, ctx)
	client, err := grpcapi.Dial(ctx, p.ready.Address, p.ready.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err := client.Create(ctx, application.CreateSession{SessionID: "main", BuildName: sample.skill, RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Apply(ctx, application.ApplyBuild{SessionID: "main", ExpectedRevision: 1, RequestID: "import", XML: string(xml)}); err != nil {
		t.Fatal(err)
	}
	connection := []string{"--address", p.ready.Address, "--token", p.ready.Token}
	m := connectMCP(t, ctx, append([]string{"--attach", "main"}, connection...)...)
	defer m.Close()
	fromMCP := structuredView(t, call(t, ctx, m, "calculate_build", map[string]any{"expected_revision": 2, "request_id": "calculate"}, false).StructuredContent)
	fromCLI := buildView(t, cli(t, ctx, 0, append([]string{"session", "get", "--session", "main"}, connection...)...))
	fromGRPC, err := client.Snapshot(ctx, "main")
	if err != nil {
		t.Fatal(err)
	}
	view, err := application.ViewSnapshot(fromGRPC)
	if err != nil || !reflect.DeepEqual(view, fromCLI) || !reflect.DeepEqual(view, fromMCP) {
		t.Fatalf("адаптеры вернули разные сборки: %v", err)
	}
	result := view.Build.Calculation
	if result.Skill != sample.skill || result.Status != "complete" || math.Abs(result.Stats["TotalDPS"]-sample.dps) > 1e-9 {
		t.Fatalf("неверный результат %s: %+v", sample.skill, result)
	}
}
