package mcpapi

import (
	"context"
	"fmt"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type renameInput struct {
	BuildName        string `json:"build_name" jsonschema:"Новое название сборки"`
	ExpectedRevision uint64 `json:"expected_revision" jsonschema:"Ревизия из последнего снимка"`
	RequestID        string `json:"request_id" jsonschema:"Уникальный идентификатор команды; сохраняйте для точного повтора"`
}

func New(commands application.Commands, sessionID string, timeout time.Duration) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "pob2_engine", Version: buildinfo.Version}, &mcp.ServerOptions{
		Instructions: fmt.Sprintf("Выбрана сессия %q. Для правок указывайте текущую ревизию и request_id. По умолчанию XML рассчитывает PoB. Поиск и допустимые правки зависят от встроенного каталога. Проверяйте status и issues в результате.", sessionID),
	})
	openWorld, destructive := false, true
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_snapshot", Description: "Получить снимок выбранной сессии и её ревизию",
		OutputSchema: calculationOutputSchema[application.SnapshotView](),
		Annotations:  &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, application.SnapshotView, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		value, err := commands.Snapshot(ctx, sessionID)
		return snapshotResult(value, err)
	})
	mcp.AddTool(server, &mcp.Tool{
		Name: "rename_build", Description: "Переименовать сборку выбранной сессии с проверкой ревизии",
		OutputSchema: calculationOutputSchema[application.SnapshotView](),
		Annotations:  &mcp.ToolAnnotations{IdempotentHint: true, DestructiveHint: &destructive, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input renameInput) (*mcp.CallToolResult, application.SnapshotView, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		value, err := commands.Rename(ctx, application.RenameBuild{SessionID: sessionID, BuildName: input.BuildName, ExpectedRevision: input.ExpectedRevision, RequestID: input.RequestID})
		return snapshotResult(value, err)
	})
	if full, ok := commands.(application.BuildCommands); ok {
		addBuildTools(server, full, sessionID, timeout)
	}
	if catalog, ok := commands.(application.CatalogCommands); ok {
		addCatalogTools(server, catalog, sessionID, timeout)
	}
	addMaintenanceTools(server, commands, timeout)
	addEditorTools(server, commands, sessionID)
	if reader, ok := commands.(application.TreeGeometryCommands); ok {
		addTreeGeometryTool(server, reader, sessionID)
	}
	if reader, ok := commands.(application.TreeTooltipCommands); ok {
		addTreeTooltipTool(server, reader, sessionID)
	}
	if reader, ok := commands.(application.BuildReportCommands); ok {
		addBuildReportTool(server, reader, sessionID)
	}
	if reader, ok := commands.(application.TreePowerCommands); ok {
		addTreePowerTool(server, reader, sessionID)
	}
	if reader, ok := commands.(application.EquipmentDetailsCommands); ok {
		addEquipmentDetailsTool(server, reader, sessionID)
	}
	if reader, ok := commands.(application.TreeTextureCommands); ok {
		addTreeTextureTool(server, reader)
	}
	return server
}
