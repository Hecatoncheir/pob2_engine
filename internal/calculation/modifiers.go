package calculation

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type Modifier struct {
	Stat   string  `json:"stat"`
	Kind   string  `json:"kind"`
	Value  float64 `json:"value"`
	Source string  `json:"source"`
}

var flatPattern = regexp.MustCompile(`^([+-]?[0-9]+(?:\.[0-9]+)?) to (.+)$`)
var percentPattern = regexp.MustCompile(`^([+-]?[0-9]+(?:\.[0-9]+)?)% (increased|reduced|more|less) (.+)$`)
var resistancePattern = regexp.MustCompile(`^([+-]?[0-9]+(?:\.[0-9]+)?)% to (.+)$`)

var flatNames = map[string][]string{
	"Strength": {"Str"}, "Dexterity": {"Dex"}, "Intelligence": {"Int"},
	"all Attributes": {"Str", "Dex", "Int"}, "maximum Life": {"Life"}, "maximum Mana": {"Mana"},
}
var percentNames = map[string]string{
	"Spell Damage": "SpellDamage", "Fire Damage": "FireDamage", "Elemental Damage": "ElementalDamage", "Projectile Damage": "ProjectileDamage",
	"Cold Damage": "ColdDamage", "Lightning Damage": "LightningDamage",
	"Damage": "Damage", "Cast Speed": "CastSpeed", "Critical Hit Chance": "CritChance", "Critical Hit Chance for Spells": "CritChance",
	"Critical Damage Bonus": "CritBonus", "maximum Life": "Life", "maximum Mana": "Mana", "Mana Regeneration Rate": "ManaRegen",
}
var resistanceNames = map[string][]string{
	"Fire Resistance": {"FireResist"}, "Cold Resistance": {"ColdResist"}, "Lightning Resistance": {"LightningResist"},
	"Chaos Resistance": {"ChaosResist"}, "all Elemental Resistances": {"FireResist", "ColdResist", "LightningResist"},
}

func modifiers(line, source string) ([]Modifier, error) {
	line = strings.TrimSpace(line)
	if matches := flatPattern.FindStringSubmatch(line); matches != nil {
		return makeModifiers(flatNames[matches[2]], "BASE", matches[1], source)
	}
	if matches := resistancePattern.FindStringSubmatch(line); matches != nil {
		return makeModifiers(resistanceNames[matches[2]], "BASE", matches[1], source)
	}
	if matches := percentPattern.FindStringSubmatch(line); matches != nil {
		stat, ok := percentNames[matches[3]]
		if !ok {
			return nil, fmt.Errorf("неподдержанный модификатор: %s", line)
		}
		kind, value := "INC", matches[1]
		if matches[2] == "reduced" || matches[2] == "less" {
			value = "-" + strings.TrimPrefix(value, "+")
		}
		if matches[2] == "more" || matches[2] == "less" {
			kind = "MORE"
		}
		if kind == "MORE" && !strings.HasSuffix(stat, "Damage") {
			return nil, fmt.Errorf("неподдержанный множитель: %s", line)
		}
		return makeModifiers([]string{stat}, kind, value, source)
	}
	return nil, fmt.Errorf("неподдержанный модификатор: %s", line)
}

func makeModifiers(stats []string, kind, raw, source string) ([]Modifier, error) {
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsInf(value, 0) || math.IsNaN(value) || math.Abs(value) > 1000000 || len(stats) == 0 {
		return nil, fmt.Errorf("неподдержанное свойство или значение %s", raw)
	}
	if kind == "MORE" && value < -100 {
		return nil, fmt.Errorf("множитель меньше -100%%")
	}
	if math.Trunc(value) != value && (kind != "BASE" || !strings.HasSuffix(stats[0], "Resist")) {
		return nil, fmt.Errorf("дробное значение свойства пока не поддерживается: %s", raw)
	}
	var result []Modifier
	for _, stat := range stats {
		result = append(result, Modifier{Stat: stat, Kind: kind, Value: value, Source: source})
	}
	return result, nil
}

func total(mods []Modifier, kind string, names ...string) float64 {
	value := 0.0
	if kind == "MORE" {
		value = 1
	}
	for _, mod := range mods {
		if mod.Kind != kind {
			continue
		}
		for _, name := range names {
			if mod.Stat != name {
				continue
			}
			if kind == "MORE" {
				value *= 1 + mod.Value/100
			} else {
				value += mod.Value
			}
		}
	}
	return value
}

func scaled(mods []Modifier, base float64, name string) float64 {
	return (base + total(mods, "BASE", name)) * math.Max(0, 1+total(mods, "INC", name)/100) * total(mods, "MORE", name)
}
