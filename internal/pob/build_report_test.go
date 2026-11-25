package pob

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestBuildReportCacheReturnsIndependentJSON(t *testing.T) {
	const xml, name = "<PathOfBuilding2/>", "Сборка"
	const saved = `{"stats":{"TotalDPS":17}}`
	calculator := &Calculator{
		gate:       make(chan struct{}, 1),
		lastReport: buildReportCache{xml: xml, name: name, report: json.RawMessage(saved)},
	}
	answer, err := calculator.readBuildReport(t.Context(), xml, name)
	if err != nil {
		t.Fatal(err)
	}
	answer[0] = '['
	if got := string(calculator.lastReport.report); got != saved {
		t.Fatalf("выдача изменила сохранённый отчёт: %s", got)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := calculator.readBuildReport(ctx, xml, name); !errors.Is(err, context.Canceled) {
		t.Fatalf("отменённый запрос получил отчёт: %v", err)
	}
	calculator.stop()
	if calculator.lastReport.report != nil {
		t.Fatal("после остановки Lua остался отчёт")
	}
}
