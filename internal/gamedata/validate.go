package gamedata

import (
	"fmt"
	"slices"
)

func (d *Dataset) validate() error {
	if len(d.catalog.Classes) == 0 || len(d.catalog.Gems) == 0 || len(d.catalog.Items) == 0 || len(d.catalog.Nodes) == 0 {
		return fmt.Errorf("неполный каталог")
	}
	if err := d.validateClasses(); err != nil {
		return err
	}
	if err := d.validateGems(); err != nil {
		return err
	}
	if err := d.validateItems(); err != nil {
		return err
	}
	if err := d.validateNodes(); err != nil {
		return err
	}
	c := d.catalog.Constants
	if !finite(c.ResistanceFloor) || c.ResistanceFloor < -1000 || c.ResistanceFloor > 0 {
		return fmt.Errorf("неверная нижняя граница сопротивлений")
	}
	for _, v := range []float64{c.LifePerLevel, c.ManaPerLevel, c.MaxResistance, c.CritBonus, c.ManaRegen} {
		if !finite(v) || v <= 0 {
			return fmt.Errorf("неверная числовая константа")
		}
	}
	if c.MaxResistance > 100 || c.ManaRegen > 1 {
		return fmt.Errorf("константа вне допустимого диапазона")
	}
	return nil
}

func (d *Dataset) validateClasses() error {
	ids, names := make(map[int]bool), make(map[string]bool)
	for _, c := range d.catalog.Classes {
		if c.ID < 1 || c.Name == "" || ids[c.ID] || names[c.Name] || c.Strength < 0 || c.Dexterity < 0 || c.Intelligence < 0 {
			return fmt.Errorf("неверный или повторный класс %d", c.ID)
		}
		ids[c.ID], names[c.Name] = true, true
		if _, ok := d.Node(c.StartNode); !ok {
			return fmt.Errorf("нет старта класса %s", c.Name)
		}
	}
	return nil
}

func (d *Dataset) validateGems() error {
	aliases := make(map[string]string)
	ids := make(map[string]bool)
	for _, g := range d.catalog.Gems {
		if ids[g.ID] {
			return fmt.Errorf("повторный камень %s", g.ID)
		}
		ids[g.ID] = true
		for _, name := range []string{g.ID, g.GameID, g.VariantID, g.SkillID, g.Name} {
			if name == "" || (aliases[name] != "" && aliases[name] != g.ID) {
				return fmt.Errorf("неверный или повторный камень %s", g.ID)
			}
			aliases[name] = g.ID
		}
		if !finite(g.CastTime) || g.CastTime < 0 || (!g.Support && g.CastTime == 0) || len(g.Levels) == 0 {
			return fmt.Errorf("нет параметров камня %s", g.ID)
		}
		if g.Support && g.Family == "" {
			return fmt.Errorf("нет семейства поддержки %s", g.ID)
		}
		if err := validateLevels(g); err != nil {
			return err
		}
		seen := make(map[string]bool)
		for _, stat := range g.ConstantStats {
			if stat.Name == "" || seen[stat.Name] || !finite(stat.Value) {
				return fmt.Errorf("неверное свойство камня %s", g.ID)
			}
			seen[stat.Name] = true
		}
	}
	return nil
}

func validateLevels(g Gem) error {
	for i, l := range g.Levels {
		if l.Level != i+1 || l.RequiredLevel < 0 || l.RequiredLevel > 100 || l.RequiredIntelligence < 0 {
			return fmt.Errorf("неверный уровень камня %s", g.ID)
		}
		for _, value := range []float64{l.FireMin, l.FireMax, l.CritChance, l.Mana} {
			if !finite(value) || value < 0 {
				return fmt.Errorf("неверное число камня %s", g.ID)
			}
		}
		if l.FireMax < l.FireMin || l.CritChance > 100 {
			return fmt.Errorf("параметр вне диапазона: %s", g.ID)
		}
	}
	return nil
}

func (d *Dataset) validateItems() error {
	names := make(map[string]bool)
	for _, i := range d.catalog.Items {
		if i.Name == "" || names[i.Name] || i.RequiredLevel < 0 || i.RequiredLevel > 100 ||
			(i.SlotType != "Ring" && i.SlotType != "Amulet") {
			return fmt.Errorf("неверная или повторная база %s", i.Name)
		}
		names[i.Name] = true
	}
	return nil
}

func (d *Dataset) validateNodes() error {
	ids := make(map[int]bool)
	for _, n := range d.catalog.Nodes {
		if n.ID < 1 || n.Name == "" || ids[n.ID] {
			return fmt.Errorf("неверный или повторный пассив %d", n.ID)
		}
		ids[n.ID] = true
		if err := d.validateEdges(n); err != nil {
			return err
		}
		for _, stat := range n.Stats {
			if stat == "" {
				return fmt.Errorf("пустое свойство пассива %d", n.ID)
			}
		}
	}
	return nil
}

func (d *Dataset) validateEdges(n Node) error {
	seen := make(map[int]bool)
	for _, id := range n.Neighbors {
		target, ok := d.Node(id)
		if !ok || id == n.ID || seen[id] || !slices.Contains(target.Neighbors, n.ID) {
			return fmt.Errorf("повреждена связь пассива %d → %d", n.ID, id)
		}
		seen[id] = true
	}
	return nil
}
