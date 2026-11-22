-- Читает справочные данные уже загруженного PoB. Исходники установки не меняются.
local catalog = {}

local function gemEntry(id, gem)
	local effect = assert(gem.grantedEffect)
	local levels, qualities = {}, {}
	for level in pairs(effect.levels) do
		if type(level) == "number" and level >= 1 and level <= 40 then
			levels[#levels + 1] = { level = level }
		end
	end
	table.sort(levels, function(a, b) return a.level < b.level end)
	for quality = 0, 100 do qualities[#qualities + 1] = quality end
	return { id = id, game_id = gem.gameId, variant_id = gem.variantId,
		name = gem.name, skill_id = gem.grantedEffectId, support = effect.support == true,
		family = effect.gemFamily and effect.gemFamily[1] or "",
		mode = effect.statSets and effect.statSets[1] and effect.statSets[1].label or "",
		levels = levels, qualities = qualities }
end

local function gems()
	local entries = {}
	for id, gem in pairs(data.gems) do
		if gem.grantedEffect then entries[#entries + 1] = gemEntry(id, gem) end
	end
	return entries
end

local function classes(tree)
	local entries = {}
	for id, class in pairs(tree.classes) do
		entries[#entries + 1] = { id = id, name = class.name,
			strength = class.base_str, dexterity = class.base_dex, intelligence = class.base_int,
			start_node = class.startNodeId }
	end
	return entries
end

local function items()
	local entries = {}
	for name, base in pairs(data.itemBases) do
		if not base.hidden then
			entries[#entries + 1] = { name = name, slot_type = base.type,
				implicit = base.implicit or "", required_level = base.req and base.req.level or 0 }
		end
	end
	return entries
end

local function nodes(tree)
	local entries = {}
	for id, node in pairs(tree.nodes) do
		if type(id) == "number" and not node.isOnlyImage then
			entries[#entries + 1] = { id = node.id, name = node.dn or node.name or tostring(id),
				stats = node.sd or {}, neighbors = node.linkedId or {} }
		end
	end
	return entries
end

function catalog.read()
	local tree = assert(main:LoadTree(latestTreeVersion))
	return { tree_version = latestTreeVersion, classes = classes(tree),
		gems = gems(), items = items(), nodes = nodes(tree) }
end

return catalog
