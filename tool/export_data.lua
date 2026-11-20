-- Извлечение числовых данных из закреплённого PoB; алгоритмы остаются отдельно.
dofile(POB_MCP_PACKAGE_DIR .. "/lua/headless_bootstrap.lua")
local json = require("dkjson")
local file = assert(io.open("TreeData/0_5/tree.json", "rb"))
local tree = assert(json.decode(file:read("*a")))
file:close()
local result = { classes = {}, gems = {}, items = {}, nodes = {}, constants = {} }
for _, class in ipairs(tree.classes) do
	local start
	for id, node in pairs(tree.nodes) do
		for _, name in ipairs(node.classesStart or {}) do
			if name == class.name then start = tonumber(id) end
		end
	end
	result.classes[#result.classes + 1] = { id = class.integerId, name = class.name,
		strength = class.base_str, dexterity = class.base_dex, intelligence = class.base_int, start_node = start }
end
local effects = { FireballPlayer = true, SupportControlledDestructionPlayer = true, SupportRapidCastingPlayer = true }
for id, gem in pairs(data.gems) do
	if effects[gem.grantedEffectId] and id:match("^Metadata/Items/Gems/") then
		local effect = data.skills[gem.grantedEffectId]
		local exported = { id = id, game_id = gem.gameId, variant_id = gem.variantId, name = gem.name, skill_id = gem.grantedEffectId, support = effect.support == true,
			cast_time = effect.castTime or 0, family = effect.gemFamily and effect.gemFamily[1] or "",
			levels = {}, constant_stats = effect.statSets[1].constantStats or {} }
		local maximum = effect.support and 1 or 20
		for level = 1, maximum do
			local values = effect.levels[level]
			local stats = effect.statSets[1].levels[level] or effect.statSets[1].levels[1]
			exported.levels[#exported.levels + 1] = { level = level, required_level = values.levelRequirement or 0,
				required_intelligence = calcLib.getGemStatRequirement(values.levelRequirement or 0, gem.reqInt or 0, effect.support),
				fire_min = effect.support and 0 or stats[1], fire_max = effect.support and 0 or stats[2],
				crit_chance = values.critChance or 0, mana = values.cost and values.cost.Mana or 0 }
		end
		result.gems[#result.gems + 1] = exported
	end
end
for _, name in ipairs({ "Lazuli Ring", "Ruby Ring", "Pearl Ring", "Prismatic Ring", "Lapis Amulet" }) do
	local base = assert(data.itemBases[name])
	result.items[#result.items + 1] = { name = name, slot_type = base.type, implicit = base.implicit or "",
		required_level = base.req and base.req.level or 0 }
end
local selected = { [54447] = true, [4739] = true, [18845] = true, [55807] = true, [6686] = true, [51184] = true }
for _, class in ipairs(result.classes) do selected[class.start_node] = true end
for id in pairs(selected) do
	local node = assert(tree.nodes[tostring(id)])
	local neighbors = {}
	for _, edge in ipairs(node.connections or {}) do
		if selected[edge.id] then neighbors[#neighbors + 1] = edge.id end
	end
	result.nodes[#result.nodes + 1] = { id = id, name = node.name, stats = node.stats or {}, neighbors = neighbors }
end
for _, key in ipairs({ "life_per_level", "mana_per_level", "base_maximum_all_resistances_%", "base_critical_hit_damage_bonus" }) do
	result.constants[key] = assert(data.characterConstants[key])
end
result.constants.mana_regeneration_base = assert(data.misc.ManaRegenBase)
result.constants.resistance_floor = assert(data.misc.ResistFloor)
io.stdout:write(assert(json.encode(result)), "\n")
