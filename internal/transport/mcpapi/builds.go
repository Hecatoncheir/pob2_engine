package mcpapi

import (
	"context"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type importInput struct {
	XML              string `json:"xml" jsonschema:"XML PathOfBuilding2 до 1 МиБ"`
	ExpectedRevision uint64 `json:"expected_revision"`
	RequestID        string `json:"request_id"`
}
type editInput struct {
	Edit             build.Edit `json:"edit" jsonschema:"Ровно одна правка: level, config, gems, equipment, passives, selection, add_support или tree_node"`
	ExpectedRevision uint64     `json:"expected_revision"`
	RequestID        string     `json:"request_id"`
}
type calculateInput struct {
	ExpectedRevision uint64 `json:"expected_revision"`
	RequestID        string `json:"request_id"`
}
type compareInput struct {
	XML              string      `json:"xml,omitempty"`
	Edit             *build.Edit `json:"edit,omitempty"`
	ExpectedRevision uint64      `json:"expected_revision"`
}

func snapshotResult(value domain.Snapshot, err error) (*mcp.CallToolResult, application.SnapshotView, error) {
	if err != nil {
		return nil, application.SnapshotView{}, err
	}
	view, err := application.ViewSnapshot(value)
	return nil, view, err
}

func addBuildTools(server *mcp.Server, commands application.BuildCommands, session string, timeout time.Duration) {
	openWorld, destructive := false, true
	write := &mcp.ToolAnnotations{IdempotentHint: true, DestructiveHint: &destructive, OpenWorldHint: &openWorld}
	read := &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &openWorld}
	mcp.AddTool(server, &mcp.Tool{Name: "import_build", Description: "Импортировать XML в выбранную сессию с сохранением неизвестного содержимого", Annotations: write, OutputSchema: calculationOutputSchema[application.SnapshotView]()},
		func(ctx context.Context, _ *mcp.CallToolRequest, input importInput) (*mcp.CallToolResult, application.SnapshotView, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			value, err := commands.Apply(ctx, application.ApplyBuild{SessionID: session, ExpectedRevision: input.ExpectedRevision, RequestID: input.RequestID, XML: input.XML})
			return snapshotResult(value, err)
		})
	mcp.AddTool(server, &mcp.Tool{Name: "edit_build", Description: "Изменить уровень, условие, камни, экипировку, пассивы или выбранные наборы; вернуть изменённую сборку", Annotations: write, InputSchema: buildInputSchema[editInput](), OutputSchema: calculationOutputSchema[application.SnapshotView]()},
		func(ctx context.Context, _ *mcp.CallToolRequest, input editInput) (*mcp.CallToolResult, application.SnapshotView, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			value, err := commands.Apply(ctx, application.ApplyBuild{SessionID: session, ExpectedRevision: input.ExpectedRevision, RequestID: input.RequestID, Edit: &input.Edit})
			return snapshotResult(value, err)
		})
	mcp.AddTool(server, &mcp.Tool{Name: "calculate_build", Description: "Рассчитать и сохранить результат текущей сборки. Обязательно проверить build.calculation.status и issues", Annotations: write, OutputSchema: calculationOutputSchema[application.SnapshotView]()},
		func(ctx context.Context, req *mcp.CallToolRequest, input calculateInput) (*mcp.CallToolResult, application.SnapshotView, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			stop := func() {}
			if req.Params.GetProgressToken() != nil {
				ctx, stop = notifyProgress(ctx, req)
				defer stop()
			}
			value, err := commands.Calculate(ctx, application.CalculateBuild{SessionID: session, ExpectedRevision: input.ExpectedRevision, RequestID: input.RequestID})
			return snapshotResult(value, err)
		})
	mcp.AddTool(server, &mcp.Tool{Name: "compare_build", Description: "Сравнить текущую сборку с XML или одной правкой без изменения сессии; проверить статусы обоих вариантов", Annotations: read, InputSchema: buildInputSchema[compareInput](), OutputSchema: calculationOutputSchema[application.BuildComparison]()},
		func(ctx context.Context, _ *mcp.CallToolRequest, input compareInput) (*mcp.CallToolResult, application.BuildComparison, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			value, err := commands.Compare(ctx, application.CompareBuild{SessionID: session, ExpectedRevision: input.ExpectedRevision, XML: input.XML, Edit: input.Edit})
			return nil, value, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "export_build", Description: "Получить сохранённый XML и SHA-256 выбранной сессии", Annotations: read},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, application.ExportedBuild, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			value, err := commands.Export(ctx, session)
			return nil, value, err
		})
}

func notifyProgress(ctx context.Context, req *mcp.CallToolRequest) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	ctx, updates := application.WithProgress(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		stages := map[string]float64{"queued": 0, "waiting": 0, "running": 1, "saving": 2, "completed": 3}
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-updates:
				callCtx, stop := context.WithTimeout(ctx, 250*time.Millisecond)
				_ = req.Session.NotifyProgress(callCtx, &mcp.ProgressNotificationParams{ProgressToken: req.Params.GetProgressToken(), Progress: stages[update.Stage], Total: 3, Message: update.Stage})
				stop()
			}
		}
	}()
	return ctx, func() { cancel(); <-done }
}
