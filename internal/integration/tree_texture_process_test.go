package integration_test

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLuaTreeTextureAcrossProcesses(t *testing.T) {
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	t.Setenv("POB2_ENGINE_CALCULATOR", "lua")
	s := newBuildScenarioWithTimeout(t, "tree-specs/warrior-witch.xml", 90*time.Second)
	tree, err := s.client.TreeGeometry(s.ctx, "main", s.revision, true)
	if err != nil {
		t.Fatal(err)
	}
	checkTreeScene(t, tree.Scene)
	for _, format := range []string{"BC1", "BC7"} {
		query := sceneTexture(t, tree.Scene, format)
		compareTreeTexture(t, s, query)
		compareTextureBatch(t, s, query)
	}
	_, err = s.client.TreeTexture(s.ctx, build.TreeTextureQuery{File: "TreeData/../Settings.png"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatal("разрешён путь вне каталога текстур", err)
	}
}

func compareTextureBatch(t *testing.T, s *groupScenario, query build.TreeTextureQuery) {
	t.Helper()
	sizes := []uint32{query.TargetSize, 80}
	layers := []*pb.TreeTextureLayer{
		{Index: query.Index, TargetSize: sizes[0]},
		{Index: query.Index, TargetSize: sizes[1]},
	}
	stream, err := s.client.TreeTextures(s.ctx, query.File, layers)
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range sizes {
		query.TargetSize = size
		single, err := s.client.TreeTexture(s.ctx, query)
		if err != nil {
			t.Fatal(err)
		}
		batch, err := stream.Recv()
		if err != nil || !bytes.Equal(single.PNG, batch.GetTexture().GetPng()) || single.MipLevel != batch.GetTexture().GetMipLevel() {
			t.Fatalf("слой размера %d отличается от одиночного запроса: %v", size, err)
		}
	}
	if _, err := stream.Recv(); err != io.EOF {
		t.Fatalf("пакет вернул лишний слой или ошибку: %v", err)
	}
}

func sceneTexture(t *testing.T, scene *build.TreeScene, format string) build.TreeTextureQuery {
	t.Helper()
	for _, command := range scene.Commands {
		if strings.HasSuffix(command.File, format+".dds.zst") {
			return build.TreeTextureQuery{File: command.File, Index: command.Index, TargetSize: 40}
		}
	}
	t.Fatal("в исходной сцене нет текстуры", format)
	return build.TreeTextureQuery{}
}

func compareTreeTexture(t *testing.T, s *groupScenario, query build.TreeTextureQuery) {
	t.Helper()
	texture, err := s.client.TreeTexture(s.ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	info, err := png.DecodeConfig(bytes.NewReader(texture.PNG))
	if err != nil || info.Width != int(texture.Width) || info.Height != int(texture.Height) {
		t.Fatal("движок не вернул действительный PNG", err)
	}
	args := []string{"tree-texture", "--file", query.File, "--index", fmt.Sprint(query.Index), "--target-size", fmt.Sprint(query.TargetSize)}
	fromCLI := decodeCatalogJSON[build.TreeTexture](t, cli(t, s.ctx, 0, append(args, s.connection...)...))
	if !reflect.DeepEqual(texture, fromCLI) {
		t.Fatal("CLI и gRPC вернули разные PNG")
	}
	answer := call(t, s.ctx, s.mcp, "get_tree_texture", map[string]any{
		"file": query.File, "index": query.Index, "target_size": query.TargetSize,
	}, false)
	for _, content := range answer.Content {
		if asset, ok := content.(*mcp.ImageContent); ok && asset.MIMEType == "image/png" && bytes.Equal(asset.Data, texture.PNG) {
			return
		}
	}
	t.Fatal("MCP не передал PNG той же текстуры")
}
