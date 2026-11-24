package mcpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"github.com/Hecatoncheir/pob2_engine/internal/workbench"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const editorBodyLimit = 8 << 20

type localEditor struct {
	endpoint string
	client   *http.Client
	gate     chan struct{}
	token    string
	revision *int
}

func NewLocalEditor(endpoint string) (*mcp.Server, error) {
	tools, err := workbench.Manifest()
	if err != nil {
		return nil, err
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "pob2_engine", Version: buildinfo.Version}, &mcp.ServerOptions{
		Instructions: "Команды работают в разрешённом окне PoB II. Получите ревизию из pob_editor_status перед правкой.",
	})
	editor := &localEditor{endpoint: endpoint, client: localEditorHTTP(), gate: make(chan struct{}, 1)}
	for _, tool := range tools {
		if err := editor.register(server, tool); err != nil {
			return nil, err
		}
	}
	return server, nil
}

func localEditorHTTP() *http.Client {
	transport := &http.Transport{DialContext: (&net.Dialer{Timeout: 5 * time.Second}).DialContext}
	return &http.Client{
		Transport: transport, Timeout: 60 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

func (e *localEditor) register(server *mcp.Server, tool workbench.Tool) error {
	if !editorSupports(tool.Name) {
		return nil
	}
	schema, err := tool.Input.Resolve(nil)
	if err != nil {
		return fmt.Errorf("схема %s: %w", tool.Name, err)
	}
	server.AddTool(&mcp.Tool{Name: tool.Name, Description: tool.Description, InputSchema: tool.Input}, func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return e.handle(ctx, tool.Name, schema, request.Params.Arguments), nil
	})
	return nil
}

func editorSupports(name string) bool {
	return strings.HasPrefix(name, "pob_editor_") || workbench.Supports(name) || name == "pob_preview_support" || name == "pob_render_image"
}

func (e *localEditor) handle(ctx context.Context, name string, schema *jsonschema.Resolved, raw json.RawMessage) *mcp.CallToolResult {
	args, err := workbench.Arguments(schema, raw)
	if err != nil {
		return workbenchFailure("INVALID_ARGUMENTS", err)
	}
	select {
	case e.gate <- struct{}{}:
		defer func() { <-e.gate }()
	case <-ctx.Done():
		return workbenchFailure("TIMEOUT", ctx.Err())
	}
	reply, err := e.call(ctx, name, args)
	if err != nil {
		return workbenchFailure("EDITOR_UNAVAILABLE", err)
	}
	return localEditorResult(reply)
}

type editorEndpoint struct {
	Port  uint16 `json:"port"`
	Token string `json:"token"`
}

type editorRequest struct {
	Tool      string         `json:"tool"`
	Arguments map[string]any `json:"arguments"`
	Revision  *int           `json:"revision"`
}

type editorResponse struct {
	pob.WorkbenchResult
	Error   *pob.WorkbenchError `json:"error"`
	Content []struct {
		Type     string `json:"type"`
		MimeType string `json:"mimeType"`
		Data     []byte `json:"data"`
	} `json:"content"`
}

func (e *localEditor) call(ctx context.Context, name string, arguments map[string]any) (editorResponse, error) {
	endpoint, err := e.readEndpoint()
	if err != nil {
		e.revision = nil
		return editorResponse{}, err
	}
	if endpoint.Token != e.token {
		e.token, e.revision = endpoint.Token, nil
	}
	reply, err := e.post(ctx, endpoint, editorRequest{Tool: name, Arguments: arguments, Revision: e.revision})
	if err != nil {
		e.revision = nil
		return reply, err
	}
	if reply.Error != nil {
		return reply, reply.Error
	}
	e.revision = &reply.Revision
	return reply, nil
}

func (e *localEditor) readEndpoint() (editorEndpoint, error) {
	var endpoint editorEndpoint
	file, err := os.Open(e.endpoint)
	if err != nil {
		return endpoint, disabledEditor()
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, 64<<10))
	if err != nil || json.Unmarshal(raw, &endpoint) != nil || endpoint.Port < 1024 || !validEditorToken(endpoint.Token) {
		return endpoint, disabledEditor()
	}
	return endpoint, nil
}

func validEditorToken(token string) bool {
	if len(token) < 32 || len(token) > 128 {
		return false
	}
	for _, character := range token {
		if character < 33 || character > 126 {
			return false
		}
	}
	return true
}

func disabledEditor() error {
	return &pob.WorkbenchError{Code: "EDITOR_DISABLED", Message: "Включите правки через MCP в PoB II"}
}

func (e *localEditor) post(ctx context.Context, endpoint editorEndpoint, value editorRequest) (editorResponse, error) {
	var reply editorResponse
	body, err := json.Marshal(value)
	if err != nil {
		return reply, err
	}
	address := "http://127.0.0.1:" + strconv.Itoa(int(endpoint.Port)) + "/edit"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, address, bytes.NewReader(body))
	if err != nil {
		return reply, err
	}
	request.Header.Set("Authorization", "Bearer "+endpoint.Token)
	request.Header.Set("Content-Type", "application/json")
	response, err := e.client.Do(request)
	if err != nil {
		return reply, fmt.Errorf("нет ответа PoB II: %w", err)
	}
	defer response.Body.Close()
	return readEditorResponse(response)
}

func readEditorResponse(response *http.Response) (editorResponse, error) {
	var reply editorResponse
	if response.StatusCode == http.StatusUnauthorized {
		return reply, disabledEditor()
	}
	if response.StatusCode != http.StatusOK {
		return reply, fmt.Errorf("PoB II вернул HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, editorBodyLimit+1))
	if err != nil || len(body) > editorBodyLimit {
		return reply, fmt.Errorf("не удалось прочитать ответ PoB II")
	}
	if err := json.Unmarshal(body, &reply); err != nil {
		return reply, fmt.Errorf("неверный ответ PoB II: %w", err)
	}
	if reply.Result == nil && reply.Error == nil {
		return reply, fmt.Errorf("в ответе PoB II нет результата")
	}
	return reply, nil
}

func localEditorResult(reply editorResponse) *mcp.CallToolResult {
	content, err := json.Marshal(reply.WorkbenchResult)
	if err != nil {
		return workbenchFailure("INVALID_RESPONSE", err)
	}
	result := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(content)}}, StructuredContent: reply.WorkbenchResult,
	}
	for _, extra := range reply.Content {
		if extra.Type == "image" {
			result.Content = append(result.Content, &mcp.ImageContent{MIMEType: extra.MimeType, Data: extra.Data})
		}
	}
	return result
}
