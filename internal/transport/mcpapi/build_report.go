package mcpapi

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type buildReportInput struct {
	ExpectedRevision uint64 `json:"expected_revision,omitempty" jsonschema:"Ноль читает последний снимок сессии"`
}

type buildReportOutput struct {
	Revision uint64         `json:"revision"`
	Report   map[string]any `json:"report"`
}

func addBuildReportTool(server *mcp.Server, reader application.BuildReportCommands, id string) {
	openWorld := false
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_build_report",
		Description: "Прочитать полный отчёт PoB об умениях, экипировке, условиях и предупреждениях выбранной сборки",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input buildReportInput) (*mcp.CallToolResult, buildReportOutput, error) {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		view, err := reader.BuildReport(ctx, id, input.ExpectedRevision)
		if err != nil {
			return nil, buildReportOutput{}, err
		}
		result := buildReportOutput{Revision: view.Revision}
		err = json.Unmarshal(view.Report, &result.Report)
		return nil, result, err
	})
}
