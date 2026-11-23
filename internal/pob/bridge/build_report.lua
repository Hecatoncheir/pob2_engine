-- Формат подробного отчёта сохраняет контракт PoB II из побочного MCP
-- (pob-poe2-mcp 312b337fa4a24f5ef55ebd38c8fd7bd46c7de78a, MIT).
local function scalars(values)
	local result = setmetatable({}, { __jsontype = "object" })
	for key, value in pairs(values or {}) do
		if type(key) == "string" then
			local kind = type(value)
			if kind == "string" or kind == "boolean" or kind == "number" and math.abs(value) < math.huge then
				result[key] = value
			end
		end
	end
	return result
end

local function selector(control, clean, multiple)
	if multiple and (not control or not control:IsShown() or #control.list < 2) then return nil end
	local entries = {}
	for id, value in ipairs(control.list or {}) do
		entries[#entries + 1] = { id = id, label = clean(type(value) == "table" and value.label or value) }
	end
	return { entries = entries, selected = control.selIndex, enabled = control:IsShown() and control:IsEnabled() }
end

local function character(build, clean)
	local controls, spec = build.controls, build.spec
	return { name = build.buildName, level = build.characterLevel, class = spec.curClassName,
		ascendancy = spec.curAscendClassName, groups = selector(controls.mainSocketGroup, clean),
		skills = selector(controls.mainSkill, clean), parts = selector(controls.mainSkillPart, clean, true),
		statSets = selector(controls.statSet, clean, true) }
end

local function supportScope(build, group)
	if not group.source or group.source == "" then return nil end
	if group.noSupports or not group.gemList[1] then return "none" end
	local tab = build.skillsTab
	if tab.skillSets[tab.activeSkillSetId].removedSocketGroupList then return "skill" end
	return group.sourceItem and group.slot and "slot" or "none"
end

local function isSupport(gem)
	return gem.gemData and gem.gemData.grantedEffect.support == true or false
end

local function isSlotGroup(group)
	if group.source or not group.slot or not group.enabled then return false end
	for _, gem in ipairs(group.gemList) do
		if not isSupport(gem) then return false end
	end
	return true
end

local function effectiveGems(build, group, scope)
	if scope ~= "slot" then return group.gemList end
	for _, container in ipairs(build.skillsTab.socketGroupList) do
		if container.slot == group.slot and isSlotGroup(container) then
			local gems = { group.gemList[1] }
			for _, gem in ipairs(container.gemList) do gems[#gems + 1] = gem end
			return gems
		end
	end
	return group.gemList
end

local function skillGroups(build, clean)
	local groups = {}
	for id, group in ipairs(build.skillsTab.socketGroupList) do
		local gems, scope = {}, supportScope(build, group)
		for index, gem in ipairs(effectiveGems(build, group, scope)) do
			gems[#gems + 1] = { sourceIndex = index, id = gem.gemId or gem.gemData and gem.gemData.id,
				skillId = gem.skillId, name = gem.nameSpec, level = gem.level, quality = gem.quality,
				enabled = gem.enabled, support = isSupport(gem), error = gem.errMsg }
		end
		groups[#groups + 1] = { id = id, label = clean(group.displayLabel), customLabel = group.label or "",
			source = group.source, enabled = group.enabled, gems = gems, supportScope = scope,
			slotSupports = isSlotGroup(group) or nil }
	end
	return groups
end

local function unsupportedItemMods(item)
	local lines = {}
	for _, list in ipairs({ item.implicitModLines or {}, item.explicitModLines or {}, item.enchantModLines or {}, item.buffModLines or {} }) do
		for _, line in ipairs(list) do
			if line.extra or not line.modList or #line.modList == 0 then
				lines[#lines + 1] = { line = line.line, extra = line.extra }
			end
		end
	end
	return lines
end

local function sortByID(rows)
	table.sort(rows, function(a, b) return a.id < b.id end)
	return rows
end

local function equipment(tab)
	local items, slots = {}, {}
	for id, item in pairs(tab.items) do
		items[#items + 1] = { id = id, name = item.title or item.name, base = item.baseName,
			rarity = item.rarity, raw = item:BuildRaw(), unsupported = unsupportedItemMods(item) }
	end
	for _, slot in ipairs(tab.orderedSlots) do
		slots[#slots + 1] = { name = slot.slotName, itemId = slot.selItemId, active = slot.active,
			visible = slot:IsShown() == true, nodeId = slot.nodeId, weaponSet = slot.weaponSet, label = slot.label }
	end
	return sortByID(items), slots
end

local function customMods(tab)
	local blocks = {}
	local function append(title, text, enabled)
		if not text or text == "" then return end
		local unparsed = {}
		for line in text:gmatch("[^\r\n]+") do
			if line:match("%S") then
				local mods, extra = modLib.parseMod(line)
				if not mods or extra then unparsed[#unparsed + 1] = { line = line, extra = extra } end
			end
		end
		blocks[#blocks + 1] = { title = title, text = text, enabled = enabled, unparsed = unparsed }
	end
	append("input.customMods", tab.input.customMods, true)
	local config = tab.configSets[tab.activeConfigSetId]
	for _, block in ipairs(config and config.customModsList or {}) do append(block.title, block.text, block.enabled) end
	return blocks
end

local function namedSets(build)
	local result = {}
	local sources = { tree = build.treeTab.specList, skills = build.skillsTab.skillSets,
		items = build.itemsTab.itemSets, config = build.configTab.configSets }
	for kind, source in pairs(sources) do
		local rows = {}
		for id, set in pairs(source) do rows[#rows + 1] = { id = id, title = set.title } end
		result[kind] = sortByID(rows)
	end
	return result
end

local function conditions(clean)
	local result = {}
	for _, definition in ipairs(require("Modules.ConfigOptions")) do
		if definition.var then
			local choices = {}
			for _, choice in ipairs(definition.list or {}) do
				choices[#choices + 1] = { value = choice.val, label = clean(choice.label) }
			end
			result[#result + 1] = { id = definition.var, name = clean(definition.label), type = definition.type, choices = choices }
		end
	end
	return result
end

local function snapshot(build, clean, stats)
	local values = {}
	for _, key in ipairs({ "TotalDPS", "Life", "Mana", "EnergyShield", "Spirit", "FireResist", "ColdResist", "LightningResist", "ChaosResist" }) do
		values[key] = stats[key]
	end
	return { targetVersion = build.targetVersion, treeVersion = build.spec.treeVersion,
		output = values, buildInfo = character(build, clean) }
end

return function(build, clean)
	local items, slots = equipment(build.itemsTab)
	local stats, warnings, allocated = scalars(build.calcsTab.mainOutput), {}, {}
	for _, line in ipairs(build.controls.warnings.lines or {}) do warnings[#warnings + 1] = clean(line) end
	for id in pairs(build.spec.allocNodes) do allocated[#allocated + 1] = id end
	table.sort(allocated)
	return { source = "original-lua", snapshot = snapshot(build, clean, stats), stats = stats,
		config = scalars(build.configTab.input), configPlaceholders = scalars(build.configTab.placeholder),
		conditionDefinitions = conditions(clean),
		groups = skillGroups(build, clean), items = items, slots = slots, warnings = warnings, allocatedNodes = allocated,
		pointDisplay = clean(build.controls.pointDisplay.str), notes = build.notesTab.controls.edit.buf,
		customMods = customMods(build.configTab), sets = namedSets(build),
		activeSets = { tree = build.treeTab.activeSpec, skills = build.skillsTab.activeSkillSetId,
			items = build.itemsTab.activeItemSetId, config = build.configTab.activeConfigSetId },
		weaponSet = build.itemsTab.activeItemSet.useSecondWeaponSet and 2 or 1,
		validationScope = "Lua calculation and exposed warnings; not a guarantee of in-game legality or performance" }
end
