// Package calculation выполняет объявленный расчёт на Go без Lua и транспортов.
package calculation

import (
	"context"
	"fmt"
	"strings"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

type Result struct {
	Status      string             `json:"status"`
	DataID      string             `json:"data_id"`
	DataSHA256  string             `json:"data_sha256"`
	PoBCommit   string             `json:"pob_commit"`
	InputSHA256 string             `json:"input_sha256"`
	Skill       string             `json:"skill"`
	Mode        string             `json:"mode"`
	Stats       map[string]float64 `json:"stats"`
	Modifiers   []Modifier         `json:"modifiers"`
	Issues      []build.Issue      `json:"issues"`
}

type Engine struct{ data *gamedata.Dataset }

func New(data *gamedata.Dataset) (*Engine, error) {
	if data == nil {
		return nil, fmt.Errorf("нет проверенного набора данных")
	}
	return &Engine{data: data}, nil
}

func (e *Engine) Calculate(ctx context.Context, doc *build.Document) (Result, error) {
	if doc == nil {
		return Result{}, fmt.Errorf("нет документа сборки")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	model, err := doc.Model()
	if err != nil {
		return Result{}, err
	}
	digest, err := doc.InputSHA256()
	if err != nil {
		return Result{}, err
	}
	result := Result{Status: "complete", DataID: e.data.ID(), DataSHA256: e.data.SHA256(), PoBCommit: e.data.PoBCommit(), InputSHA256: digest,
		Issues: append([]build.Issue{}, model.Issues...), Modifiers: []Modifier{}}
	state := inputs{engine: e, model: model, result: &result, enemyResists: make(map[string]float64)}
	state.readCharacter()
	state.readConfig()
	state.readGems()
	state.readPassives()
	state.readEquipment()
	state.readCustomMods()
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	for _, issue := range result.Issues {
		if strings.HasPrefix(issue.Code, "unsupported") || issue.Code == "missing_input" {
			result.Status = "unsupported"
			return result, nil
		}
	}
	if len(result.Issues) > 0 {
		result.Status = "partial"
	}
	result.Stats = state.compute()
	if err := validateStats(result.Stats); err != nil {
		return Result{}, err
	}
	if len(result.Issues) > 0 {
		result.Status = "partial"
	}
	return result, ctx.Err()
}

type inputs struct {
	engine       *Engine
	model        build.Model
	result       *Result
	class        gamedata.Class
	gem          gamedata.GemLevel
	skill        gamedata.Gem
	castTime     float64
	cannotCrit   bool
	penalty      float64
	enemyResists map[string]float64
	items        []gamedata.Item
}

func (s *inputs) issue(code, path, message string) {
	s.result.Issues = append(s.result.Issues, build.Issue{Code: code, Path: path, Message: message})
}

func (s *inputs) addLine(line, source string) {
	if strings.TrimSpace(line) == "" {
		return
	}
	mods, err := modifiers(line, source)
	if err != nil {
		s.issue("unsupported_modifier", source, err.Error())
		return
	}
	s.result.Modifiers = append(s.result.Modifiers, mods...)
}

func (s *inputs) readCharacter() {
	if s.model.TreeVersion != s.engine.data.TreeVersion() || s.model.TargetVersion != "0_1" {
		s.issue("unsupported_version", "Build", "неподдерживаемая версия XML или дерева")
	}
	class, ok := s.engine.data.Class(s.model.ClassID)
	if !ok || class.Name != s.model.Class {
		s.issue("unsupported_class", "Build/className", "неизвестный класс или несовпадение его имени и ID")
	}
	s.class = class
	if s.model.AscendancyID != 0 || s.model.Ascendancy != "None" {
		s.issue("unsupported_ascendancy", "Build/ascendClassName", "возвышения пока не поддерживаются")
	}
}
