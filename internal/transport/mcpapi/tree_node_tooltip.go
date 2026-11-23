package mcpapi

import (
	"context"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type treeNodeTooltipInput struct {
	ExpectedRevision uint64 `json:"expected_revision,omitempty" jsonschema:"Ноль читает последний снимок сессии"`
	NodeID           uint32 `json:"node_id" jsonschema:"ID узла активной спецификации"`
	Compare          bool   `json:"compare,omitempty" jsonschema:"Показать изменение расчёта от узла"`
}

func addTreeTooltipTool(server *mcp.Server, reader application.TreeTooltipCommands, id string) {
	openWorld := false
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_tree_node_tooltip",
		Description: "Прочитать штатную подсказку, путь и влияние узла активной спецификации",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input treeNodeTooltipInput) (*mcp.CallToolResult, application.TreeNodeTooltipView, error) {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		tip, err := reader.TreeNodeTooltip(ctx, id, input.ExpectedRevision, input.NodeID, input.Compare)
		return nil, tip, err
	})
}
