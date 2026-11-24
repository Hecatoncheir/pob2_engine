package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestLuaWorkbenchAcrossGRPCAndMCP(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	workbench, token, closeWorker := startWorkbenchGRPC(t, ctx)
	if _, err := workbench.ListWorkbenchTools(ctx, &pb.ListWorkbenchToolsRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("запрос без токена: %v", err)
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	mcpClient := openMCP(t, ctx, "workbench")
	xml, err := os.ReadFile(filepath.Join("..", "..", "testdata", "reference", "fireball.xml"))
	if err != nil {
		t.Fatal(err)
	}
	compareWorkbenchCall(t, ctx, workbench, mcpClient, "pob_import_xml", map[string]any{"xml": string(xml)})
	compareWorkbenchCall(t, ctx, workbench, mcpClient, "pob_check_build", nil)
	compareWorkbenchCall(t, ctx, workbench, mcpClient, "pob_export_xml", nil)
	compareWorkbenchCall(t, ctx, workbench, mcpClient, "pob_render_image", map[string]any{"view": "overview", "width": 880, "height": 640})
	closeWorker()
}

func startWorkbenchGRPC(t *testing.T, ctx context.Context) (pb.WorkbenchServiceClient, string, func()) {
	t.Helper()
	cmd := exec.CommandContext(ctx, binary, "workbench", "--grpc")
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})
	var ready grpcapi.WorkbenchReady
	if err := json.NewDecoder(output).Decode(&ready); err != nil {
		t.Fatal(err)
	}
	channel, err := grpc.NewClient(ready.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = channel.Close() })
	return pb.NewWorkbenchServiceClient(channel), ready.Token, func() {
		t.Helper()
		_ = input.Close()
		if err := cmd.Wait(); err != nil {
			t.Fatalf("рабочий процесс не завершился после EOF: %v", err)
		}
	}
}

func compareWorkbenchCall(t *testing.T, ctx context.Context, client pb.WorkbenchServiceClient, tools *mcp.ClientSession, name string, arguments map[string]any) {
	t.Helper()
	input, err := json.Marshal(arguments)
	if err != nil {
		t.Fatal(err)
	}
	reply, err := client.CallWorkbenchTool(ctx, &pb.CallWorkbenchToolRequest{Name: name, ArgumentsJson: input})
	if err != nil || reply.GetError() != nil {
		t.Fatalf("gRPC %s: %v / %v", name, err, reply.GetError())
	}
	mcpReply := call(t, ctx, tools, name, arguments, false)
	type payload struct {
		Revision uint64         `json:"revision"`
		Result   map[string]any `json:"result"`
	}
	other := structuredCatalogJSON[payload](t, mcpReply.StructuredContent)
	result := decodeCatalogJSON[map[string]any](t, reply.GetResultJson())
	if name == "pob_export_xml" {
		result["xml"] = workbenchBuildModel(t, result["xml"])
		other.Result["xml"] = workbenchBuildModel(t, other.Result["xml"])
	}
	if reply.GetRevision() != other.Revision || !reflect.DeepEqual(result, other.Result) {
		t.Fatalf("ответы рабочей копии через gRPC и MCP различаются для %s", name)
	}
	if name == "pob_render_image" {
		if len(mcpReply.Content) == 0 {
			t.Fatal("MCP не вернул изображение")
		}
		image, _ := mcpReply.Content[len(mcpReply.Content)-1].(*mcp.ImageContent)
		if image == nil || !bytes.HasPrefix(reply.GetPng(), []byte("\x89PNG\r\n\x1a\n")) || !bytes.Equal(reply.GetPng(), image.Data) {
			t.Fatal("gRPC и MCP вернули разные изображения")
		}
	}
}

func workbenchBuildModel(t *testing.T, value any) build.Model {
	t.Helper()
	xml, ok := value.(string)
	if !ok {
		t.Fatal("нет экспортированного XML")
	}
	document, err := build.Parse([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	model, err := document.Model()
	if err != nil {
		t.Fatal(err)
	}
	return model
}
