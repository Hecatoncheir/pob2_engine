package workbench

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"github.com/Hecatoncheir/pob2_engine/internal/poblegacy"
	"github.com/google/jsonschema-go/jsonschema"
)

const LostCode = "WORKBENCH_LOST"

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

type Tool struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Input       *jsonschema.Schema `json:"inputSchema"`
}

type Response struct {
	Result pob.WorkbenchResult
	PNG    []byte
}

type Service struct {
	session      *pob.Workbench
	preview      *preparedPreview
	installation pob.Installation
	assets       *workbenchAssets
	gate         chan struct{}
	tools        []Tool
	schemas      map[string]*jsonschema.Resolved
	revision     int
}

func Manifest() ([]Tool, error) {
	raw, err := poblegacy.Read("data/tools.json")
	if err != nil {
		return nil, err
	}
	var tools []Tool
	err = json.Unmarshal(raw, &tools)
	return tools, err
}

func New(installation pob.Installation) (*Service, error) {
	manifest, err := Manifest()
	if err != nil {
		return nil, err
	}
	assets, err := newWorkbenchAssets(installation)
	if err != nil {
		return nil, err
	}
	s := &Service{installation: installation, assets: assets, gate: make(chan struct{}, 1), schemas: make(map[string]*jsonschema.Resolved)}
	for _, tool := range manifest {
		if !Supports(tool.Name) {
			continue
		}
		schema, err := tool.Input.Resolve(nil)
		if err != nil {
			return nil, fmt.Errorf("схема %s: %w", tool.Name, err)
		}
		s.tools = append(s.tools, tool)
		s.schemas[tool.Name] = schema
	}
	return s, nil
}

func (s *Service) Tools() []Tool { return s.tools }

// Prepare нужен MCP перед публикацией сервера. Мастерская gRPC открывает Lua по запросу.
func (s *Service) Prepare(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	return s.openSession(ctx)
}

func (s *Service) openSession(ctx context.Context) error {
	if s.session != nil {
		return nil
	}
	session, err := pob.NewWorkbench(ctx, s.installation)
	if err == nil {
		s.session = session
	}
	return err
}

func (s *Service) Close() {
	s.gate <- struct{}{}
	defer func() { <-s.gate }()
	s.discardPreview()
	if s.session != nil {
		s.session.Close()
		s.session = nil
	}
}

func (s *Service) discardPreview() {
	if s.preview != nil {
		s.preview.worker.Close()
		s.preview = nil
	}
}

func (s *Service) finishPreview(err *error) {
	if *err != nil {
		s.discardPreview()
	}
}

func Supports(name string) bool {
	switch name {
	case "pob_check_build", "pob_gem_supports", "pob_preview_support", "pob_tree_textures", "pob_render_image":
		return true
	default:
		return workbenchMethods[name] != ""
	}
}

func (s *Service) Invoke(ctx context.Context, name string, raw json.RawMessage) (Response, error) {
	schema, ok := s.schemas[name]
	if !ok {
		return Response{}, failure("INVALID_ARGUMENTS", fmt.Errorf("неизвестный инструмент %q", name))
	}
	arguments, err := Arguments(schema, raw)
	if err != nil {
		return Response{}, failure("INVALID_ARGUMENTS", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	select {
	case s.gate <- struct{}{}:
		defer func() { <-s.gate }()
	case <-ctx.Done():
		return Response{}, failure("TIMEOUT", ctx.Err())
	}
	if err := ctx.Err(); err != nil {
		return Response{}, failure("TIMEOUT", err)
	}
	response, err := s.run(ctx, name, arguments)
	if s.session != nil && s.session.Closed() {
		return Response{}, &pob.WorkbenchError{Code: LostCode, Message: "Рабочая копия PoB потеряна. Перезапустите мастерскую и откройте сборку заново."}
	}
	if err != nil && ctx.Err() != nil {
		return Response{}, failure("TIMEOUT", ctx.Err())
	}
	return response, err
}

func (s *Service) run(ctx context.Context, name string, args map[string]any) (Response, error) {
	if name != "pob_tree_textures" {
		if err := s.openSession(ctx); err != nil {
			return Response{}, failure("CALCULATION_FAILED", err)
		}
	}
	if name == "pob_render_image" {
		return s.renderImage(ctx, args)
	}
	result, err := s.call(ctx, name, args)
	if err != nil {
		return Response{}, failure("CALCULATION_FAILED", err)
	}
	s.revision = result.Revision
	return Response{Result: result}, nil
}

func Arguments(schema *jsonschema.Resolved, raw json.RawMessage) (map[string]any, error) {
	arguments := map[string]any{}
	if len(raw) != 0 {
		if err := json.Unmarshal(raw, &arguments); err != nil {
			return nil, err
		}
	}
	return arguments, schema.Validate(arguments)
}

func failure(code string, err error) *pob.WorkbenchError {
	var original *pob.WorkbenchError
	if errors.As(err, &original) {
		return original
	}
	return &pob.WorkbenchError{Code: code, Message: err.Error()}
}

func (s *Service) call(ctx context.Context, name string, arguments map[string]any) (pob.WorkbenchResult, error) {
	switch name {
	case "pob_check_build":
		return s.checkBuild(ctx, arguments)
	case "pob_gem_supports":
		return s.compatibleSupports(ctx, arguments)
	case "pob_preview_support":
		return s.previewSupport(ctx, arguments)
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

func (s *Service) checkBuild(ctx context.Context, arguments map[string]any) (pob.WorkbenchResult, error) {
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

func workbenchResult(revision int, value any) (pob.WorkbenchResult, error) {
	content, err := json.Marshal(value)
	return pob.WorkbenchResult{Revision: revision, Result: content}, err
}
