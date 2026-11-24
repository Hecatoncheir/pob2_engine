package mcpapi

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"github.com/Hecatoncheir/pob2_engine/internal/workbench"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewWorkbench(service *workbench.Service) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "pob2_engine", Version: buildinfo.Version}, &mcp.ServerOptions{
		Instructions: "Рабочая копия сборки использует расчёты исходного Lua PoB.",
	})
	for _, tool := range service.Tools() {
		server.AddTool(&mcp.Tool{Name: tool.Name, Description: tool.Description, InputSchema: tool.Input}, func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result, err := service.Invoke(ctx, tool.Name, request.Params.Arguments)
			return workbenchReply(result, err), nil
		})
	}
	return server
}

func workbenchReply(response workbench.Response, err error) *mcp.CallToolResult {
	if err != nil {
		return workbenchFailure("CALCULATION_FAILED", err)
	}
	content, err := json.Marshal(response.Result)
	if err != nil {
		return workbenchFailure("INVALID_RESPONSE", err)
	}
	result := &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(content)}},
		StructuredContent: response.Result,
	}
	if len(response.PNG) != 0 {
		result.Content = append(result.Content, &mcp.ImageContent{MIMEType: "image/png", Data: response.PNG})
	}
	return result
}

func workbenchFailure(code string, err error) *mcp.CallToolResult {
	failure := &pob.WorkbenchError{Code: code, Message: err.Error()}
	var original *pob.WorkbenchError
	if errors.As(err, &original) {
		failure = original
	}
	content, _ := json.Marshal(failure)
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: string(content)}}}
}
