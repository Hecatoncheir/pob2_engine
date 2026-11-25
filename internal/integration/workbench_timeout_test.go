package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestLuaWorkbenchReportsLostCopyAfterTimeout(t *testing.T) {
	marker := useStallingWorkbench(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	client, token, closeWorker := startWorkbenchGRPC(t, ctx)
	defer closeWorker()
	workbenchCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	importBuild := &pb.CallWorkbenchToolRequest{
		Name: "pob_import_xml", ArgumentsJson: []byte(`{"xml":"<PathOfBuilding2/>"}`),
	}
	requireWorkbenchImport(t, workbenchCtx, client, importBuild)

	short, stop := context.WithTimeout(workbenchCtx, 2*time.Second)
	defer stop()
	reply, err := client.CallWorkbenchTool(short, &pb.CallWorkbenchToolRequest{
		Name: "pob_import_xml", ArgumentsJson: []byte(`{"xml":"<PathOfBuilding2 stall='true'/>"}`),
	})
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("ожидалось истечение срока запроса: %v / %v", err, reply.GetError())
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("Lua не начал запрос до истечения срока: %v", err)
	}

	_, err = client.CallWorkbenchTool(workbenchCtx, &pb.CallWorkbenchToolRequest{Name: "pob_export_xml"})
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("после остановки Lua ожидалась недоступность рабочей копии: %v", err)
	}
	next, nextToken, closeNext := startWorkbenchGRPC(t, ctx)
	defer closeNext()
	requireWorkbenchImport(t, metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+nextToken), next, importBuild)
}

func requireWorkbenchImport(t *testing.T, ctx context.Context, client pb.WorkbenchServiceClient, request *pb.CallWorkbenchToolRequest) {
	t.Helper()
	reply, err := client.CallWorkbenchTool(ctx, request)
	if err != nil || reply.GetError() != nil || reply.GetRevision() != 1 {
		t.Fatalf("сборка не открылась: %v / %v, ревизия %d", err, reply.GetError(), reply.GetRevision())
	}
}

func useStallingWorkbench(t *testing.T) string {
	t.Helper()
	if os.Getenv("POB2_ENGINE_POB_DIR") == "" {
		t.Skip("для проверки Lua укажите POB2_ENGINE_POB_DIR")
	}
	manager, err := pob.Local()
	if err != nil {
		t.Fatal(err)
	}
	installation, err := manager.Installed()
	if err != nil {
		t.Fatal(err)
	}
	layout, err := installation.Layout()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	marker := filepath.Join(root, "request-started")
	writeWorkbenchFixture(t, root, "Launch.lua", []byte(stallingWorkbench))
	writeWorkbenchFixture(t, root, "TreeData/0_5/tree.json", []byte(`{"assets":{}}`))
	copyWorkbenchFixture(t, layout.Runtime, root, "lua/dkjson.lua")
	if runtime.GOOS == "windows" {
		copyWorkbenchFixture(t, layout.Runtime, root, "lua51.dll")
	}
	t.Setenv("POB2_ENGINE_POB_DIR", root)
	t.Setenv("POB2_ENGINE_POB_NATIVE_DIR", layout.Native)
	t.Setenv("POB2_TEST_WORKBENCH_STARTED", filepath.ToSlash(marker))
	return marker
}

func copyWorkbenchFixture(t *testing.T, source, target, name string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(source, name))
	if err != nil {
		t.Fatal(err)
	}
	writeWorkbenchFixture(t, target, name, data)
}

func writeWorkbenchFixture(t *testing.T, root, name string, data []byte) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

// Заглушка оставляет настоящий Lua-процесс занятым после начала запроса.
const stallingWorkbench = `
launch = {}
common = { xml = { ParseXML = function() return {{ elem = "PathOfBuilding2" }} end } }
local control = { IsShown = function() return false end, IsEnabled = function() return true end }
local fixture = { calcsTab = { mainOutput = {} }, spec = {}, treeTab = {}, controls = {
	mainSocketGroup = control, mainSkill = control,
} }
function fixture:SaveDB() return "<PathOfBuilding2/>" end
__mainObject__ = { main = { modes = { BUILD = fixture } } }
function __mainObject__.main:SetMode(_, _, _, xml)
	if xml:find("stall", 1, true) then
		io.output(os.getenv("POB2_TEST_WORKBENCH_STARTED")):close()
		while true do end
	end
end
`
