package mcpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type editorCaller interface {
	CallEditor(context.Context, string, string, string) (string, error)
}

var editorDescriptions = []struct{ name, description string }{
	{"pob_editor_status", "Проверить подключение окна PoB II и получить ревизию редактора"},
	{"pob_editor_ui", "Управлять разделами, диалогами и элементами окна PoB II"},
	{"pob_editor_app", "Управлять приложением PoB II"},
	{"pob_editor_components", "Просматривать и обновлять компоненты PoB II"},
	{"pob_editor_library", "Работать с библиотекой сборок PoB II"},
	{"pob_editor_history", "Отменить или повторить правку в PoB II"},
	{"pob_editor_notes", "Читать и изменять заметки к сборке в PoB II"},
	{"pob_editor_files", "Импортировать или экспортировать файлы и коды PoB II"},
	{"pob_editor_progression", "Работать с этапами прокачки PoB II"},
	{"pob_editor_job", "Получить состояние фоновой задачи PoB II"},
}

func addEditorTools(server *mcp.Server, commands any, session string) {
	caller, _ := commands.(editorCaller)
	schema := json.RawMessage(`{"type":"object","properties":{"action":{"type":"string"},"expectedRevision":{"type":"integer","minimum":0}},"additionalProperties":true}`)
	for _, tool := range editorDescriptions {
		server.AddTool(&mcp.Tool{Name: tool.name, Description: tool.description, InputSchema: schema},
			func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return callEditorTool(ctx, caller, session, tool.name, request.Params.Arguments), nil
			})
	}
}

func callEditorTool(ctx context.Context, caller editorCaller, session, tool string, arguments json.RawMessage) *mcp.CallToolResult {
	if caller == nil {
		return editorToolError(fmt.Errorf("для управления окном подключитесь к сессии через mcp --attach"))
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	if len(arguments) == 0 {
		arguments = json.RawMessage("{}")
	}
	raw, err := caller.CallEditor(ctx, session, tool, string(arguments))
	if err != nil {
		return editorToolError(err)
	}
	result, err := editorToolResult(raw)
	if err != nil {
		return editorToolError(err)
	}
	return result
}

type editorEnvelope struct {
	Result   map[string]any `json:"result"`
	Revision int            `json:"revision"`
	Error    *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Content []struct {
		Type     string `json:"type"`
		MIMEType string `json:"mimeType"`
		Data     []byte `json:"data"`
	} `json:"content"`
}

func editorToolResult(raw string) (*mcp.CallToolResult, error) {
	var reply editorEnvelope
	if err := json.Unmarshal([]byte(raw), &reply); err != nil {
		return nil, err
	}
	if reply.Error != nil {
		return nil, fmt.Errorf("%s: %s", reply.Error.Code, reply.Error.Message)
	}
	if reply.Result == nil {
		return nil, fmt.Errorf("редактор не вернул результат")
	}
	reply.Result["revision"] = reply.Revision
	text, err := json.Marshal(reply.Result)
	if err != nil {
		return nil, err
	}
	result := &mcp.CallToolResult{
		StructuredContent: reply.Result,
		Content:           []mcp.Content{&mcp.TextContent{Text: string(text)}},
	}
	for _, item := range reply.Content {
		if item.Type == "image" && item.MIMEType == "image/png" {
			result.Content = append(result.Content, &mcp.ImageContent{MIMEType: item.MIMEType, Data: item.Data})
		}
	}
	return result, nil
}

func editorToolError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}
}
