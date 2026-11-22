package mcpapi

import (
	"context"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type previewSupportInput struct {
	SupportID        string `json:"support_id" jsonschema:"Точный ID или псевдоним поддержки из каталога"`
	Level            int    `json:"level,omitempty" jsonschema:"Уровень поддержки; 0 означает 1"`
	Quality          int    `json:"quality,omitempty" jsonschema:"Качество поддержки, 0–100"`
	ExpectedRevision uint64 `json:"expected_revision"`
}

func addCatalogTools(server *mcp.Server, commands application.CatalogCommands, session string, timeout time.Duration) {
	openWorld := false
	read := &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld}
	mcp.AddTool(server, &mcp.Tool{Name: "get_catalog", Description: "Найти поддержанные камни, классы, базы предметов или пассивы; обязательно учитывать coverage; limit до 100", Annotations: read},
		func(ctx context.Context, _ *mcp.CallToolRequest, input application.CatalogQuery) (*mcp.CallToolResult, application.CatalogPage, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			value, err := commands.Catalog(ctx, input)
			return nil, value, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "preview_support", Description: "Предпросмотр поддержки с выбранными уровнем и качеством в полной группе без записи; проверить compatible, status и issues обоих вариантов", Annotations: read, OutputSchema: calculationOutputSchema[application.SupportPreview]()},
		func(ctx context.Context, _ *mcp.CallToolRequest, input previewSupportInput) (*mcp.CallToolResult, application.SupportPreview, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			value, err := commands.PreviewSupport(ctx, application.PreviewSupport{SessionID: session, ExpectedRevision: input.ExpectedRevision, SupportID: input.SupportID, Level: input.Level, Quality: input.Quality})
			return nil, value, err
		})
}
