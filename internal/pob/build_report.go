package pob

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (c *Calculator) ReadBuildReport(ctx context.Context, doc *build.Document, name string) (json.RawMessage, error) {
	xml, err := doc.ExportWithoutProgression()
	if err != nil {
		return nil, err
	}
	answer, err := c.run(ctx, &request{Operation: "build_report", XML: pobXML.Replace(string(xml)), BuildName: name})
	if err != nil {
		return nil, err
	}
	if answer.Error != "" {
		return nil, fmt.Errorf("PoB Lua: %s", answer.Error)
	}
	if len(answer.BuildReport) == 0 {
		return nil, fmt.Errorf("PoB не вернул подробный отчёт")
	}
	return answer.BuildReport, nil
}
