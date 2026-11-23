package mcpapi

import (
	"context"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type treeTextureOutput struct {
	Width    uint32 `json:"width"`
	Height   uint32 `json:"height"`
	MipLevel uint32 `json:"mip_level"`
}

func addTreeTextureTool(server *mcp.Server, reader application.TreeTextureCommands) {
	openWorld := false
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_tree_texture",
		Description: "Получить PNG текстуры из команды отрисовки дерева PoB",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input build.TreeTextureQuery) (*mcp.CallToolResult, treeTextureOutput, error) {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		texture, err := reader.TreeTexture(ctx, input)
		if err != nil {
			return nil, treeTextureOutput{}, err
		}
		result := &mcp.CallToolResult{Content: []mcp.Content{&mcp.ImageContent{Data: texture.PNG, MIMEType: "image/png"}}}
		return result, treeTextureOutput{Width: texture.Width, Height: texture.Height, MipLevel: texture.MipLevel}, nil
	})
}
