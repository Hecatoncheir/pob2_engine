package mcpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"github.com/Hecatoncheir/pob2_engine/internal/poblegacy"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var workbenchMethods = map[string]string{
	"pob_catalog":            "getCatalog",
	"pob_create_build":       "createBuild",
	"pob_import_xml":         "loadBuild",
	"pob_import_character":   "importCharacter",
	"pob_add_skill_group":    "addSkillGroup",
	"pob_update_skill_group": "updateSkillGroup",
	"pob_delete_skill_group": "deleteSkillGroup",
	"pob_add_item":           "addItem",
	"pob_set_config":         "setConfig",
	"pob_set_custom_mods":    "setCustomMods",
	"pob_tree_action":        "treeAction",
	"pob_tree_options":       "treeOptions",
	"pob_select_main_skill":  "selectMainSkill",
	"pob_node_tooltip":       "getNodeTooltip",
	"pob_export_xml":         "exportBuild",
	"pob_get_tree":           "getTreeGeometry",
	"pob_tree_icons":         "getTreeIconAssets",
	"pob_tree_hover":         "getTreeHover",
	"pob_node_power":         "getNodePower",
	"pob_gem_icons":          "getGemIconAssets",
	"pob_gem_tooltip":        "getGemTooltip",
	"pob_inspect_item":       "inspectItem",
	"pob_compare_item":       "compareItem",
	"pob_trade_catalog":      "getTradeCatalog",
	"pob_trade_search":       "buildTradeSearch",
	"pob_equip_item":         "equipItem",
	"pob_select_weapon_set":  "selectWeaponSet",
}

type workbenchTool struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Input       *jsonschema.Schema `json:"inputSchema"`
}

type workbenchServer struct {
	session      *pob.Workbench
	installation pob.Installation
	assets       *workbenchAssets
	gate         chan struct{}
	revision     int
}

func NewWorkbench(session *pob.Workbench, installation pob.Installation) (*mcp.Server, error) {
	manifest, err := poblegacy.Read("data/tools.json")
	if err != nil {
		return nil, err
	}
	var tools []workbenchTool
	if err := json.Unmarshal(manifest, &tools); err != nil {
		return nil, err
	}
	assets, err := newWorkbenchAssets(installation)
	if err != nil {
		return nil, err
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "pob2_engine", Version: buildinfo.Version}, &mcp.ServerOptions{
		Instructions: "Рабочая копия сборки использует расчёты исходного Lua PoB.",
	})
	implementation := &workbenchServer{session: session, installation: installation, assets: assets, gate: make(chan struct{}, 1)}
	for _, tool := range tools {
		if err := implementation.register(server, tool); err != nil {
			return nil, err
		}
	}
	return server, nil
}

func (s *workbenchServer) register(server *mcp.Server, tool workbenchTool) error {
	if !workbenchSupports(tool.Name) {
		return nil
	}
	schema, err := tool.Input.Resolve(nil)
	if err != nil {
		return fmt.Errorf("схема %s: %w", tool.Name, err)
	}
	server.AddTool(&mcp.Tool{Name: tool.Name, Description: tool.Description, InputSchema: tool.Input}, func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return s.handle(ctx, tool.Name, schema, request.Params.Arguments), nil
	})
	return nil
}

func workbenchSupports(name string) bool {
	return workbenchMethods[name] != "" || name == "pob_check_build" || name == "pob_gem_supports" || name == "pob_tree_textures" || name == "pob_render_image"
}

func (s *workbenchServer) handle(ctx context.Context, name string, schema *jsonschema.Resolved, raw json.RawMessage) *mcp.CallToolResult {
	arguments, err := workbenchArguments(schema, raw)
	if err != nil {
		return workbenchFailure("INVALID_ARGUMENTS", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	select {
	case s.gate <- struct{}{}:
		defer func() { <-s.gate }()
	case <-ctx.Done():
		return workbenchFailure("TIMEOUT", ctx.Err())
	}
	if name == "pob_render_image" {
		return s.renderImage(ctx, arguments)
	}
	result, err := s.call(ctx, name, arguments)
	if err != nil {
		return workbenchFailure("CALCULATION_FAILED", err)
	}
	s.revision = result.Revision
	content, err := json.Marshal(result)
	if err != nil {
		return workbenchFailure("INVALID_RESPONSE", err)
	}
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: string(content)}},
		StructuredContent: result,
	}
}

func workbenchArguments(schema *jsonschema.Resolved, raw json.RawMessage) (map[string]any, error) {
	arguments := map[string]any{}
	if len(raw) != 0 {
		if err := json.Unmarshal(raw, &arguments); err != nil {
			return nil, err
		}
	}
	return arguments, schema.Validate(arguments)
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

func (s *workbenchServer) call(ctx context.Context, name string, arguments map[string]any) (pob.WorkbenchResult, error) {
	switch name {
	case "pob_check_build":
		return s.checkBuild(ctx, arguments)
	case "pob_gem_supports":
		return s.compatibleSupports(ctx, arguments)
	case "pob_tree_textures":
		result, err := s.assets.texture(ctx, arguments["token"].(string))
		result.Revision = s.revision
		return result, err
	default:
		result, err := s.session.Call(ctx, workbenchMethods[name], arguments)
		if err != nil {
			return result, err
		}
		return s.assets.prepare(ctx, name, result)
	}
}

func (s *workbenchServer) checkBuild(ctx context.Context, arguments map[string]any) (pob.WorkbenchResult, error) {
	report, err := s.session.Call(ctx, "getBuildReport", nil)
	if err != nil {
		return report, err
	}
	checks, passed, err := workbenchChecks(report.Result, arguments["targets"])
	if err != nil {
		return report, err
	}
	return workbenchResult(report.Revision, map[string]any{
		"report": report.Result, "revision": report.Revision, "checks": checks, "targetsPassed": passed,
	})
}

func workbenchChecks(report json.RawMessage, rawTargets any) ([]map[string]any, any, error) {
	checks := []map[string]any{}
	targets, _ := rawTargets.([]any)
	if len(targets) == 0 {
		return checks, nil, nil
	}
	var summary struct {
		Stats map[string]any `json:"stats"`
	}
	if err := json.Unmarshal(report, &summary); err != nil {
		return nil, nil, err
	}
	all := true
	for _, raw := range targets {
		target := raw.(map[string]any)
		actual := summary.Stats[target["stat"].(string)]
		passed := meetsTarget(actual, target)
		checks = append(checks, map[string]any{"stat": target["stat"], "actual": actual, "target": target, "passed": passed})
		all = all && passed
	}
	return checks, all, nil
}

func meetsTarget(actual any, target map[string]any) bool {
	value, ok := actual.(float64)
	if !ok {
		return false
	}
	if minimum, ok := target["min"].(float64); ok && value < minimum {
		return false
	}
	maximum, limited := target["max"].(float64)
	return !limited || value <= maximum
}

func (s *workbenchServer) compatibleSupports(ctx context.Context, arguments map[string]any) (pob.WorkbenchResult, error) {
	export, err := s.session.Call(ctx, "exportBuild", nil)
	if err != nil {
		return export, err
	}
	var document struct {
		XML string `json:"xml"`
	}
	if err := json.Unmarshal(export.Result, &document); err != nil {
		return export, err
	}
	preview, err := pob.NewWorkbench(ctx, s.installation)
	if err != nil {
		return export, err
	}
	defer preview.Close()
	if _, err := preview.Call(ctx, "loadBuild", map[string]any{"xml": document.XML}); err != nil {
		return export, err
	}
	return readWorkbenchSupports(ctx, preview, export.Revision, arguments)
}

func readWorkbenchSupports(ctx context.Context, preview *pob.Workbench, revision int, arguments map[string]any) (pob.WorkbenchResult, error) {
	operation := "addSkillGroup"
	if arguments["groupId"] != nil {
		operation = "updateSkillGroup"
	}
	edited, err := preview.Call(ctx, operation, arguments)
	if err != nil {
		return edited, err
	}
	var group struct {
		ID int `json:"groupId"`
	}
	if err := json.Unmarshal(edited.Result, &group); err != nil {
		return edited, err
	}
	supports, err := preview.Call(ctx, "getCompatibleSupports", map[string]any{"groupId": group.ID})
	supports.Revision = revision
	return supports, err
}

func workbenchResult(revision int, value any) (pob.WorkbenchResult, error) {
	content, err := json.Marshal(value)
	return pob.WorkbenchResult{Revision: revision, Result: content}, err
}
