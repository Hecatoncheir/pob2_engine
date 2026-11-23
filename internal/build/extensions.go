package build

import (
	"math"
	"strconv"
	"strings"
)

type layout struct{ attributes, children string }

// Известные поля перечислены явно: расширения сохраняются, но не теряются в расчёте.
var layouts = map[string]layout{
	"PathOfBuilding2":    {"", "Build Items Tree Skills Config Calcs Notes TreeView Party Import"},
	"Build":              {"level className ascendClassName targetVersion mainSocketGroup mainSkillIndex characterLevelAutoMode viewMode", "PlayerStat MinionStat FullDPSSkill Buffs TimelessData"},
	"Buffs":              {"buffList combatList curseList", ""},
	"TimelessData":       {"searchList searchListFallback devotionVariant1 devotionVariant2 socketFilterDistance", ""},
	"Import":             {"exportParty", ""},
	"Items":              {"activeItemSet useSecondWeaponSet showStatDifferences", "Item ItemSet TradeSearchWeights"},
	"TradeSearchWeights": {"", "Stat"},
	"Stat":               {"label stat weightMult", ""},
	"Item":               {"id", "ModRange"},
	"ModRange":           {"id range", ""},
	"ItemSet":            {"id title useSecondWeaponSet", "Slot"},
	"Slot":               {"name itemId itemPbURL", ""},
	"Tree":               {"activeSpec", "Spec"},
	"Spec":               {"treeVersion secondaryAscendClassId classInternalId classId ascendancyInternalId ascendClassId nodes masteryEffects title", "URL Sockets Overrides"},
	"URL":                {"", ""},
	"Sockets":            {"", "Socket"},
	"Socket":             {"nodeId itemId", ""},
	"Overrides":          {"", "AttributeOverride"},
	"AttributeOverride":  {"strNodes dexNodes intNodes", ""},
	"Skills":             {"activeSkillSet defaultGemLevel defaultGemQuality sortGemsByDPSField sortGemsByDPS showSupportGemTypes showLegacyGems", "SkillSet"},
	"SkillSet":           {"id title", "Skill"},
	"Skill":              {"enabled mainActiveSkill mainActiveSkillCalcs includeInFullDPS label", "Gem"},
	"Gem":                {"quality corruptLevel corrupted level gemId enabled enableGlobal1 enableGlobal2 statSetIndex statSetIndexCalcs count nameSpec skillId variantId", ""},
	"Config":             {"activeConfigSet", "ConfigSet"},
	"ConfigSet":          {"id title", "Input Placeholder"},
	"Input":              {"name boolean number string", ""},
	"Placeholder":        {"name number", ""},
	"Party":              {"ShowAdvanceTools append destination", ""},
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
	inspectElement(node, parent, model, active)
	retainInactiveIssues(model, firstIssue, active)
}

func inspectElement(node *element, parent string, model *Model, active bool) {
	name := node.name.Local
	path := parent + "/" + name
	if isNonCalculationSection(name) {
		return
	}
	schema, known := layouts[name]
	if !known || node.name.Space != "" {
		model.issue(path, "неизвестная секция XML сохранена")
		return
	}
	inspectAttributes(node, path, schema.attributes, model)
	inspectSpecialValues(node, path, model)
	inspectChildren(node, path, schema.children, model, active)
}

func retainInactiveIssues(model *Model, first int, active bool) {
	if active {
		return
	}
	for index := first; index < len(model.Issues); index++ {
		model.Issues[index].Code = "retained_inactive_xml"
	}
}

func isNonCalculationSection(name string) bool {
	switch name {
	case "Calcs", "Notes", "TreeView", "PlayerStat", "MinionStat", "FullDPSSkill":
		return true
	default:
		return false
	}
}

func inspectAttributes(node *element, path, known string, model *Model) {
	for _, attribute := range node.attrs {
		if attribute.Name.Space != "" || !containsWord(known, attribute.Name.Local) {
			model.issue(path+"/@"+attribute.Name.Local, "неизвестный атрибут XML сохранён")
		}
	}
}

func inspectChildren(node *element, path, known string, model *Model, active bool) {
	specIndex := 0
	for _, part := range node.parts {
		child := part.element
		if child == nil {
			continue
		}
		if !containsWord(known, child.name.Local) {
			model.issue(path+"/"+child.name.Local, "неизвестное содержимое XML сохранено")
			continue
		}
		childActive := active
		if node.name.Local == "Tree" && child.name.Local == "Spec" {
			specIndex++
			childActive = active && specIndex == model.ActiveSets.Tree
		}
		inspectExtensions(child, path, model, childActive)
	}
}

func activeElement(node *element, model *Model) bool {
	selected := 0
	switch node.name.Local {
	case "Skill", "Gem":
		enabled, err := boolean(node.attr("enabled", "true"), "enabled")
		return err == nil && enabled
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
		inspectModRange(node, path, model)
	case "Spec":
		inspectSpecOptions(node, path, model)
	case "Socket":
		if node.attr("itemId", "0") != "0" {
			model.issue(path, "самоцвет в гнезде требует оригинального Lua PoB")
		}
	case "AttributeOverride":
		inspectAttributeOverride(node, path, model)
	case "Gem":
		inspectGemOptions(node, path, model)
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

func inspectModRange(node *element, path string, model *Model) {
	id, idError := strconv.Atoi(node.attr("id", ""))
	value, valueError := strconv.ParseFloat(node.attr("range", ""), 64)
	if idError != nil || id < 1 || valueError != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
		model.issue(path, "неверные параметры диапазона свойства предмета")
	}
}

func inspectSpecOptions(node *element, path string, model *Model) {
	if node.attr("ascendancyInternalId", "") != "" {
		model.issue(path, "внутренний идентификатор возвышения пока не поддерживается")
	}
	if node.attr("masteryEffects", "") != "" {
		model.issue(path, "эффекты мастерства пока не поддерживаются")
	}
	if value := node.attr("secondaryAscendClassId", "nil"); value != "nil" && value != "0" {
		model.issue(path, "второе возвышение пока не поддерживается")
	}
}

func inspectAttributeOverride(node *element, path string, model *Model) {
	for _, attribute := range node.attrs {
		if attribute.Value != "" {
			model.issue(path, "замена характеристик пассивов пока не поддерживается")
			return
		}
	}
}

func inspectGemOptions(node *element, path string, model *Model) {
	if node.attr("count", "1") != "1" || node.attr("corruptLevel", "0") != "0" || node.attr("enableGlobal2", "false") == "true" {
		model.issue(path, "дополнительные режимы камня пока не поддерживаются")
	}
}
