package calculation

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

var questFlags = []string{
	"questAct 1ClearfellBeira", "questAct 1FreythornKing In The Mists", "questAct 1Ogham ManorCandlemass",
	"questAct 2Spires of DesharSisters of Garukhan Shrine", "questAct 3Azak BogIgnagduk", "questAct 3Jiquani's MachinariumBlackjaw",
	"questAct 4Eye of HinekoraSilent Hall", "questInterlude 2Khari CrossingMolten Shrine", "questInterlude 3Kriar VillageLythara",
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s *inputs) readConfig() {
	s.penalty = -60
	for _, name := range questFlags {
		value, ok := s.model.Config[name]
		if !ok || value.Type != "boolean" || value.Value != "false" {
			s.issue("unsupported_quest", "Config/"+name, "для этого набора награда задания должна быть явно выключена")
		}
	}
	for _, name := range sortedKeys(s.model.Config) {
		value := s.model.Config[name]
		path := "Config/" + name
		if slices.Contains(questFlags, name) {
			continue
		}
		switch name {
		case "customMods":
			if value.Type != "string" {
				s.issue("unsupported_config", path, "нужна строка модификаторов")
			}
		case "enemyIsBoss":
			if value.Type != "string" || value.Value != "None" {
				s.issue("unsupported_enemy", path, "поддерживается обычный противник None")
			}
		case "resistancePenalty":
			s.penalty = s.number(value, path, -100, 0)
		case "enemyFireResist":
			s.enemyResist = s.number(value, path, -100, 75)
		case "enemyLevel":
			s.number(value, path, 1, 100)
		default:
			s.issue("unsupported_config", path, "неподдерживаемое условие расчёта")
		}
	}
}

func (s *inputs) number(value build.ConfigValue, path string, minimum, maximum float64) float64 {
	number, err := strconv.ParseFloat(value.Value, 64)
	if value.Type != "number" || err != nil || number < minimum || number > maximum {
		s.issue("unsupported_config", path, "значение вне поддерживаемого диапазона")
		return 0
	}
	return number
}

func (s *inputs) readGems() {
	if len(s.model.Groups) > 100 {
		s.issue("unsupported_limit", "Skills", "слишком много групп")
		return
	}
	var group *build.Group
	for i := range s.model.Groups {
		candidate := &s.model.Groups[i]
		if candidate.ID == s.model.MainGroup {
			group = candidate
		}
		if candidate.Enabled && candidate.ID != s.model.MainGroup && len(candidate.Gems) > 0 {
			s.issue("unsupported_group", fmt.Sprintf("Skills/%d", candidate.ID), "дополнительная активная группа пока не поддерживается")
		}
	}
	if group == nil || !group.Enabled {
		s.issue("missing_input", "Skills", "не выбрана включённая группа")
		return
	}
	if len(group.Gems) > 12 {
		s.issue("unsupported_limit", "Skills", "не более 12 камней в группе")
		return
	}
	if group.MainSkill != 1 {
		s.issue("unsupported_skill", "Skills/mainActiveSkill", "поддерживается первое активное умение")
	}
	activeCount := 0
	families := make(map[string]bool)
	for index, gem := range group.Gems {
		if !gem.Enabled {
			continue
		}
		path := fmt.Sprintf("Skills/%d/Gem/%d", group.ID, index+1)
		selected, ok := s.engine.data.Gem(gem.SkillID)
		if !ok {
			s.issue("unsupported_gem", path, "неизвестный камень")
			continue
		}
		id := strings.Replace(gem.ID, "Metadata/Items/Gem/", "Metadata/Items/Gems/", 1)
		if (id != "" && id != selected.ID && gem.ID != selected.GameID) || (gem.Name != "" && gem.Name != selected.Name) ||
			(gem.VariantID != "" && gem.VariantID != selected.VariantID) {
			s.issue("unsupported_gem", path, "идентификаторы камня противоречат друг другу")
		}
		if gem.Quality != 0 || gem.Corrupted || gem.StatSet != 1 || gem.Level > len(selected.Levels) {
			s.issue("unsupported_gem", path, "неподдерживаемый уровень, качество или режим камня")
			continue
		}
		if !selected.Support {
			activeCount++
			s.gem, s.castTime = selected.Levels[gem.Level-1], selected.CastTime
			continue
		}
		if families[selected.Family] {
			s.issue("unsupported_support", path, "поддержка этого семейства уже используется")
		}
		families[selected.Family] = true
		for _, stat := range selected.ConstantStats {
			switch stat.Name {
			case "base_cast_speed_+%":
				s.result.Modifiers = append(s.result.Modifiers, Modifier{"CastSpeed", "INC", stat.Value, path})
			case "support_controlled_destruction_spell_damage_+%_final":
				s.result.Modifiers = append(s.result.Modifiers, Modifier{"Damage", "MORE", stat.Value, path})
				s.cannotCrit = true
			default:
				s.issue("unsupported_support", path, "неизвестное свойство поддержки")
			}
		}
	}
	if activeCount != 1 {
		s.issue("missing_input", "Skills", "нужно одно активное умение Fireball")
	}
}

func (s *inputs) readPassives() {
	allocated := make(map[int]bool)
	for _, id := range s.model.Nodes {
		allocated[id] = true
	}
	if !allocated[s.class.StartNode] {
		s.issue("missing_input", "Tree", "не указан старт класса")
		return
	}
	if len(allocated)-1 > s.model.Level-1 {
		s.issue("unsupported_passives", "Tree", "недостаточно очков без наград заданий")
	}
	for _, id := range s.model.Nodes {
		node, ok := s.engine.data.Node(id)
		if !ok {
			s.issue("unsupported_passive", fmt.Sprintf("Tree/%d", id), "неподдерживаемый пассив")
			continue
		}
		for _, line := range node.Stats {
			s.addLine(line, fmt.Sprintf("Tree/%d", id))
		}
	}
	visited := map[int]bool{}
	queue := []int{s.class.StartNode}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if visited[id] {
			continue
		}
		visited[id] = true
		node, _ := s.engine.data.Node(id)
		for _, neighbor := range node.Neighbors {
			if allocated[neighbor] && !visited[neighbor] {
				queue = append(queue, neighbor)
			}
		}
	}
	if len(visited) != len(allocated) {
		s.issue("unsupported_passives", "Tree", "выделенные пассивы не связаны со стартом класса")
	}
}

func (s *inputs) readCustomMods() {
	value := s.model.Config["customMods"]
	if value.Type != "string" {
		return
	}
	if len(value.Value) > 65536 {
		s.issue("unsupported_limit", "Config/customMods", "пользовательские свойства превышают 65536 байт")
		return
	}
	for i, line := range strings.Split(value.Value, "\n") {
		s.addLine(line, fmt.Sprintf("Config/customMods/%d", i+1))
	}
}
