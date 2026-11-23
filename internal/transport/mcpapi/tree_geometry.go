package mcpapi

import (
	"context"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type treeGeometryInput struct {
	ExpectedRevision uint64 `json:"expected_revision,omitempty" jsonschema:"Ноль читает последний снимок сессии"`
}

func addTreeGeometryTool(server *mcp.Server, reader application.TreeGeometryCommands, id string) {
	openWorld := false
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_tree_geometry",
		Description: "Прочитать координаты, связи и назначение узлов активного дерева PoB",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input treeGeometryInput) (*mcp.CallToolResult, application.TreeGeometryView, error) {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		tree, err := reader.TreeGeometry(ctx, id, input.ExpectedRevision)
		return nil, tree, err
	})
}
