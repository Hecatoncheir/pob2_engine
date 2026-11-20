package build

import (
	"math"
	"strconv"
	"strings"
)

type layout struct{ attributes, children string }

// Известные поля перечислены явно: расширения сохраняются, но не теряются в расчёте.
var layouts = map[string]layout{
	"PathOfBuilding2":   {"", "Build Items Tree Skills Config Calcs Notes TreeView Party Import"},
	"Build":             {"level className ascendClassName targetVersion mainSocketGroup characterLevelAutoMode viewMode", "PlayerStat MinionStat FullDPSSkill Buffs TimelessData"},
	"Buffs":             {"buffList combatList curseList", ""},
	"TimelessData":      {"searchList searchListFallback devotionVariant1 devotionVariant2", ""},
	"Import":            {"exportParty", ""},
	"Items":             {"activeItemSet useSecondWeaponSet showStatDifferences", "Item ItemSet"},
	"Item":              {"id", "ModRange"},
	"ModRange":          {"id range", ""},
	"ItemSet":           {"id title useSecondWeaponSet", "Slot"},
	"Slot":              {"name itemId itemPbURL", ""},
	"Tree":              {"activeSpec", "Spec"},
	"Spec":              {"treeVersion secondaryAscendClassId classInternalId classId ascendancyInternalId ascendClassId nodes masteryEffects title", "URL Sockets Overrides"},
	"URL":               {"", ""},
	"Sockets":           {"", ""},
	"Overrides":         {"", "AttributeOverride"},
	"AttributeOverride": {"strNodes dexNodes intNodes", ""},
	"Skills":            {"activeSkillSet defaultGemLevel defaultGemQuality sortGemsByDPSField sortGemsByDPS showSupportGemTypes showLegacyGems", "SkillSet"},
	"SkillSet":          {"id title", "Skill"},
	"Skill":             {"enabled mainActiveSkill mainActiveSkillCalcs includeInFullDPS label", "Gem"},
	"Gem":               {"quality corruptLevel corrupted level gemId enabled enableGlobal1 enableGlobal2 statSetIndex statSetIndexCalcs count nameSpec skillId variantId", ""},
	"Config":            {"activeConfigSet", "ConfigSet"},
	"ConfigSet":         {"id title", "Input Placeholder"},
	"Input":             {"name boolean number string", ""},
	"Placeholder":       {"name number", ""},
	"Party":             {"ShowAdvanceTools append destination", ""},
}

func containsWord(words, value string) bool {
	for _, word := range strings.Fields(words) {
		if word == value {
			return true
		}
	}
	return false
}

func (d *Document) findExtensions(m *Model) {
	inspectExtensions(d.root, "", m, true)
}

func inspectExtensions(node *element, parent string, model *Model, active bool) {
	active = active && activeElement(node, model)
	firstIssue := len(model.Issues)
	defer func() {
		if active {
			return
		}
		for i := firstIssue; i < len(model.Issues); i++ {
			model.Issues[i].Code = "retained_inactive_xml"
		}
	}()
	name := node.name.Local
	path := parent + "/" + name
	if name == "Calcs" || name == "Notes" || name == "TreeView" || name == "PlayerStat" || name == "MinionStat" || name == "FullDPSSkill" {
		return
	}
	schema, known := layouts[name]
	if !known || node.name.Space != "" {
		model.issue(path, "неизвестная секция XML сохранена")
		return
	}
	for _, attr := range node.attrs {
		if attr.Name.Space != "" || !containsWord(schema.attributes, attr.Name.Local) {
			model.issue(path+"/@"+attr.Name.Local, "неизвестный атрибут XML сохранён")
		}
	}
	inspectSpecialValues(node, path, model)
	specIndex := 0
	for _, part := range node.parts {
		if part.element == nil {
			continue
		}
		if !containsWord(schema.children, part.element.name.Local) {
			model.issue(path+"/"+part.element.name.Local, "неизвестное содержимое XML сохранено")
			continue
		}
		childActive := active
		if name == "Tree" && part.element.name.Local == "Spec" {
			specIndex++
			childActive = active && specIndex == model.ActiveSets.Tree
		}
		inspectExtensions(part.element, path, model, childActive)
	}
}

func activeElement(node *element, model *Model) bool {
	selected := 0
	switch node.name.Local {
	case "SkillSet":
		selected = model.ActiveSets.Skills
	case "ItemSet":
		selected = model.ActiveSets.Items
	case "ConfigSet":
		selected = model.ActiveSets.Config
	case "Item":
		for _, id := range model.Slots {
			if node.attr("id", "") == strconv.Itoa(id) {
				return true
			}
		}
		return false
	default:
		return true
	}
	return node.attr("id", "1") == strconv.Itoa(selected)
}

func inspectSpecialValues(node *element, path string, model *Model) {
	switch node.name.Local {
	case "ModRange":
		id, err := strconv.Atoi(node.attr("id", ""))
		value, numberErr := strconv.ParseFloat(node.attr("range", ""), 64)
		if err != nil || id < 1 || numberErr != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
			model.issue(path, "неверные параметры диапазона свойства предмета")
		}
	case "Spec":
		if node.attr("masteryEffects", "") != "" {
			model.issue(path, "эффекты мастерства пока не поддерживаются")
		}
		if value := node.attr("secondaryAscendClassId", "nil"); value != "nil" && value != "0" {
			model.issue(path, "второе возвышение пока не поддерживается")
		}
	case "AttributeOverride":
		for _, attr := range node.attrs {
			if attr.Value != "" {
				model.issue(path, "замена характеристик пассивов пока не поддерживается")
				break
			}
		}
	case "Gem":
		if node.attr("count", "1") != "1" || node.attr("corruptLevel", "0") != "0" || node.attr("enableGlobal2", "false") == "true" {
			model.issue(path, "дополнительные режимы камня пока не поддерживаются")
		}
	case "Build":
		if node.attr("characterLevelAutoMode", "false") == "true" {
			model.issue(path, "автоматический уровень персонажа пока не поддерживается")
		}
	case "Party":
		if node.attr("append", "false") != "false" {
			model.issue(path, "внешние эффекты группы пока не поддерживаются")
		}
	}
}
