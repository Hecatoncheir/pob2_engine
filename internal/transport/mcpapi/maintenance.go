package mcpapi

import (
	"context"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/maintenance"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type updateInput struct {
	Apply  bool   `json:"apply,omitempty" jsonschema:"Явно разрешить обновление или откат. По умолчанию только проверка"`
	Target string `json:"target,omitempty" jsonschema:"Цель all, engine, pob, pob@vX.Y.Z или pob-rollback. По умолчанию all"`
}

func addMaintenanceTools(server *mcp.Server, target any, timeout time.Duration) {
	commands, ok := target.(maintenance.Commands)
	if !ok {
		commands = maintenance.Local()
	}
	openWorld, destructive := true, true
	mcp.AddTool(server, &mcp.Tool{
		Name: "update", Description: "Проверить обновление или возможность отката PoB. При apply=true установить проверенный пакет либо выбрать предыдущую установку для следующего запуска.",
		Annotations: &mcp.ToolAnnotations{OpenWorldHint: &openWorld, DestructiveHint: &destructive},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input updateInput) (*mcp.CallToolResult, maintenance.UpdateResult, error) {
		ctx, cancel := context.WithTimeout(ctx, max(timeout, maintenance.UpdateTimeout))
		defer cancel()
		result, err := commands.Update(ctx, input.Apply, input.Target)
		return nil, result, err
	})
	mcp.AddTool(server, &mcp.Tool{
		Name: "doctor", Description: "Проверить состояние процесса, установленные версии и доступные обновления PoB2Engine и Path of Building",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, maintenance.Report, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		result, err := commands.Doctor(ctx)
		return nil, result, err
	})
}
