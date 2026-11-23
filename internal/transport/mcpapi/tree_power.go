package mcpapi

import (
	"context"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type treePowerInput struct {
	ExpectedRevision uint64 `json:"expected_revision,omitempty" jsonschema:"Ноль читает последний снимок сессии"`
	build.TreePowerQuery
}

func addTreePowerTool(server *mcp.Server, reader application.TreePowerCommands, id string) {
	openWorld := false
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_tree_power",
		Description: "Найти влияние пассивов на выбранный показатель или прочитать каталог показателей без metric",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input treePowerInput) (*mcp.CallToolResult, application.TreePowerView, error) {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		report, err := reader.TreePower(ctx, id, input.ExpectedRevision, input.TreePowerQuery)
		return nil, report, err
	})
}
