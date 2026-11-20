package mcpapi

import (
	"context"
	"fmt"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type renameInput struct {
	BuildName        string `json:"build_name" jsonschema:"Новое название сборки"`
	ExpectedRevision uint64 `json:"expected_revision" jsonschema:"Ревизия из последнего снимка"`
	RequestID        string `json:"request_id" jsonschema:"Уникальный идентификатор команды; сохраняйте для точного повтора"`
}

func New(commands application.Commands, sessionID string, timeout time.Duration) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "pob2_engine", Version: buildinfo.Version}, &mcp.ServerOptions{
		Instructions: fmt.Sprintf("Выбрана сессия %q. Доступны снимок и переименование; расчёты и управление Flutter ещё не реализованы.", sessionID),
	})
	openWorld, destructive := false, true
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_snapshot", Description: "Получить снимок выбранной сессии и её ревизию",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, domain.Snapshot, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		value, err := commands.Snapshot(ctx, sessionID)
		return nil, value, err
	})
	mcp.AddTool(server, &mcp.Tool{
		Name: "rename_build", Description: "Переименовать сборку выбранной сессии с проверкой ревизии",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true, DestructiveHint: &destructive, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input renameInput) (*mcp.CallToolResult, domain.Snapshot, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		value, err := commands.Rename(ctx, application.RenameBuild{SessionID: sessionID, BuildName: input.BuildName, ExpectedRevision: input.ExpectedRevision, RequestID: input.RequestID})
		return nil, value, err
	})
	return server
}
