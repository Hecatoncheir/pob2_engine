package pob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

// Хранит последний готовый отчёт, доступ к нему защищает очередь Lua.
type buildReportCache struct {
	xml, name string
	report    json.RawMessage
}

func (c *Calculator) ReadBuildReport(ctx context.Context, doc *build.Document, name string) (json.RawMessage, error) {
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return nil, err
	}
	return c.readBuildReport(ctx, pobXML.Replace(string(xml)), name)
}

func (c *Calculator) readBuildReport(ctx context.Context, xml, name string) (json.RawMessage, error) {
	release, err := c.acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if cached := c.lastReport; cached.xml == xml && cached.name == name && len(cached.report) > 0 {
		return bytes.Clone(cached.report), nil
	}
	answer, err := c.requestBuildReport(ctx, xml, name)
	if err != nil {
		return nil, err
	}
	if answer.Error != "" {
		return nil, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if len(answer.BuildReport) == 0 {
		return nil, fmt.Errorf("PoB не вернул подробный отчёт")
	}
	c.lastReport = buildReportCache{xml: xml, name: name, report: bytes.Clone(answer.BuildReport)}
	return answer.BuildReport, nil
}

func (c *Calculator) requestBuildReport(ctx context.Context, xml, name string) (response, error) {
	query := &request{Operation: "build_report", XML: xml, BuildName: name,
		ReuseTreeReport: c.treeReportXML == xml && name != "~~temp~~"}
	answer, err := c.execute(ctx, query)
	if err == nil && answer.Error != "" && query.ReuseTreeReport {
		query.ReuseTreeReport = false
		return c.execute(ctx, query)
	}
	return answer, err
}
