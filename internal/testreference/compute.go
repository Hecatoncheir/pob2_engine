package testreference

// Ограниченный перенос порядка вычислений PoB:
// PathOfBuildingCommunity/PathOfBuilding-PoE2, commit 7d6f530cbdab20389ff8bc6ba97a37ac27f74e41.
// Modules/CalcSetup.lua, CalcPerform.lua, CalcDefence.lua, CalcOffence.lua, Common.lua.
// Copyright (c) 2016 David Gowor. MIT; исходный текст в .licenses/reference/pob-LICENSE.md.
// Изменено: типизированные входы, явное покрытие, отсутствие Lua и проверка неподдержанных механик.

import (
	"fmt"
	"math"
)

func round(value float64, decimals int) float64 {
	factor := math.Pow10(decimals)
	return math.Floor(value*factor+0.5) / factor
}

func (s *inputs) compute() map[string]float64 {
	mods := s.result.Modifiers
	c := s.engine.data.Constants()
	str := math.Max(0, round(scaled(mods, float64(s.class.Strength), "Str"), 0))
	dex := math.Max(0, round(scaled(mods, float64(s.class.Dexterity), "Dex"), 0))
	intelligence := math.Max(0, round(scaled(mods, float64(s.class.Intelligence), "Int"), 0))
	life := math.Max(1, round(scaled(mods, 16+c.LifePerLevel*float64(s.model.Level)+2*str, "Life"), 0))
	mana := math.Max(1, round(scaled(mods, 30+c.ManaPerLevel*float64(s.model.Level)+2*intelligence, "Mana"), 0))
	regen := math.Max(0, round(scaled(mods, mana*c.ManaRegen, "ManaRegen"), 1))
	chance := math.Min(100, math.Max(0, scaled(mods, s.gem.CritChance, "CritChance")))
	multiplier := 1 + math.Max(0, scaled(mods, c.CritBonus, "CritBonus"))/100
	if s.cannotCrit {
		chance, multiplier = 0, 0
	}
	effect := 1 + chance/100*(multiplier-1)
	speed := round(math.Max(0, 1+total(mods, "INC", "CastSpeed")/100), 2) / s.castTime
	names := []string{"Damage", "SpellDamage", s.skill.DamageType + "Damage", "ElementalDamage", "ProjectileDamage"}
	damageScale := math.Max(0, 1+total(mods, "INC", names...)/100) * total(mods, "MORE", names...)
	minimum, maximum := round(s.gem.DamageMin*damageScale, 0), round(s.gem.DamageMax*damageScale, 0)
	storedAverage := (minimum + maximum) / 2
	enemyMultiplier := 1 - s.enemyResists[s.skill.DamageType]/100
	hitAverage := storedAverage * enemyMultiplier
	averageHit := hitAverage * effect
	stats := map[string]float64{
		"Str": str, "Dex": dex, "Int": intelligence, "Life": life, "Mana": mana, "ManaRegen": regen,
		"ManaCost": s.gem.Mana, "ManaPerSecondCost": s.gem.Mana * speed, "CastRate": speed, "Speed": speed,
		"CritChance": chance, "CritMultiplier": multiplier, "CritEffect": effect, "AverageHit": averageHit, "TotalDPS": averageHit * speed,
	}
	for name, value := range map[string]float64{
		"MinBase": s.gem.DamageMin, "MaxBase": s.gem.DamageMax, "HitAverage": hitAverage, "CritAverage": hitAverage * multiplier,
		"StoredHitMin": minimum, "StoredHitMax": maximum, "StoredHitAvg": storedAverage, "StoredCombinedAvg": storedAverage * effect,
	} {
		stats[s.skill.DamageType+name] = value
	}
	for _, name := range []string{"FireResist", "ColdResist", "LightningResist", "ChaosResist"} {
		base := s.penalty
		if name == "ChaosResist" {
			base = 0
		}
		stats[name] = math.Max(c.ResistanceFloor, math.Min(c.MaxResistance, math.Trunc(base+total(mods, "BASE", name))))
	}
	if intelligence < float64(s.gem.RequiredIntelligence) || s.model.Level < s.gem.RequiredLevel {
		s.issue("requirements_not_met", "Skills/"+s.skill.Name, fmt.Sprintf("не выполнены требования %s к уровню или Intelligence", s.skill.Name))
	}
	for _, item := range s.items {
		if s.model.Level < item.RequiredLevel {
			s.issue("requirements_not_met", "Items/"+item.Name, "не выполнено требование предмета к уровню")
		}
	}
	return stats
}

func validateStats(stats map[string]float64) error {
	for _, name := range sortedKeys(stats) {
		if math.IsNaN(stats[name]) || math.IsInf(stats[name], 0) {
			return fmt.Errorf("некорректный результат %s", name)
		}
	}
	return nil
}
