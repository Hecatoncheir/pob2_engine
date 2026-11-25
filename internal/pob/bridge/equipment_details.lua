-- Описание и сравнение из pob-poe2-mcp, 312b337fa4a24f5ef55ebd38c8fd7bd46c7de78a, MIT.
local function unsupported(item)
	local result = {}
	for _, list in ipairs({ item.implicitModLines or {}, item.explicitModLines or {}, item.enchantModLines or {}, item.buffModLines or {} }) do
		for _, line in ipairs(list) do
			if line.extra or not line.modList or #line.modList == 0 then result[#result + 1] = { line = line.line, extra = line.extra } end
		end
	end
	return result
end
local function tooltip(build, item, slot)
	local collector = { lines = {} }
	function collector:AddLine(size, text)
		if type(text) == "string" and text:match("%S") and not text:find("Tip: ", 1, true) then self.lines[#self.lines + 1] = { size = size, text = text } end
	end
	function collector:AddSeparator()
		if #self.lines > 0 and not self.lines[#self.lines].separator then self.lines[#self.lines + 1] = { separator = true } end
	end
	build.itemsTab:AddItemTooltip(collector, item, slot, false)
	if collector.lines[#collector.lines] and collector.lines[#collector.lines].separator then table.remove(collector.lines) end
	return collector.lines
end
local function inspectItem(build, item, slot)
	local slots = {}
	for _, slot in ipairs(build.itemsTab.orderedSlots) do
		if not slot.nodeId and build.itemsTab:IsItemValidForSlot(item, slot.slotName) then slots[#slots + 1] = slot.slotName end
	end
	return { source = "original-lua", item = { id = item.id, name = item.name, base = item.baseName,
		rarity = item.rarity, raw = item.raw, unsupported = unsupported(item) }, slots = slots,
		compatible = slot and build.itemsTab:IsItemValidForSlot(item, slot.slotName) == true or false,
		lines = tooltip(build, item, slot) }
end
local function finite(value)
	return type(value) == "number" and value == value and math.abs(value) ~= math.huge
end
local function comparisonRow(build, stat, actor, before, after)
	local old, current = before[stat.stat], after[stat.stat]
	if not finite(old) and not finite(current) then return end
	local collector = { text = nil }
	function collector:AddLine(_, text) if text ~= "" then self.text = text end end
	if finite(old) and finite(current) then build:CompareStatList(collector, { stat }, actor, before, after, "") end
	local change = collector.text
	local benefit
	if change then
		benefit = change:find(colorCodes.POSITIVE, 1, true) ~= nil
		local first, last = change:find(" " .. stat.label, 1, true)
		if first then change = change:sub(1, first - 1) .. change:sub(last + 1) end
	end
	return { id = stat.stat, label = stat.label, before = finite(old) and old or nil, after = finite(current) and current or nil,
		beforeText = finite(old) and build:FormatStat(stat, old, before[stat.overCapStat]) or nil,
		afterText = finite(current) and build:FormatStat(stat, current, after[stat.overCapStat]) or nil,
		changeText = change, benefit = benefit }
end
local function comparisonRows(build, stats, actor, before, after, prefix)
	local rows, seen = {}, {}
	local active = actor.mainSkill and actor.mainSkill.activeEffect
	local skillFlags = active and active.statSet and active.statSet.skillFlags or {}
	for _, stat in ipairs(stats) do
		if stat.stat and not stat.childStat and not seen[stat.stat] and stat.stat ~= "SkillDPS" and
			(not stat.flag or skillFlags[stat.flag]) then
			local row = comparisonRow(build, stat, actor, before, after)
			if row then row.actor = prefix; rows[#rows + 1] = row; seen[stat.stat] = true end
		end
	end
	return rows
end
local attributeRequirements = { ReqStr = true, ReqDex = true, ReqInt = true }
local function requirementWarnings(stats, output)
	local warnings = {}
	for _, stat in ipairs(stats) do
		local value = output[stat.stat]
		if attributeRequirements[stat.stat] and stat.warnFunc and value and
			(not stat.condFunc or stat.condFunc(value, output)) then
			warnings[#warnings + 1] = stat.warnFunc(value, output)
		end
	end
	return warnings
end
local function compareItem(build, item, slot, result)
	local activeSet = build.itemsTab.activeItemSet.useSecondWeaponSet and 2 or 1
	if slot.weaponSet and slot.weaponSet ~= activeSet then
		result.comparison = { rows = {}, reason = "Выберите этот набор оружия для примерки" }
		return result
	end
	local calculate, before = build.calcsTab:GetMiscCalculator()
	local after = calculate({ repSlotName = slot.slotName, repItem = item })
	local rows = comparisonRows(build, build.displayStats, build.calcsTab.mainEnv.player, before, after, "player")
	if before.Minion and after.Minion and build.calcsTab.mainEnv.minion then
		for _, row in ipairs(comparisonRows(build, build.minionDisplayStats, build.calcsTab.mainEnv.minion, before.Minion, after.Minion, "minion")) do rows[#rows + 1] = row end
	end
	result.comparison = { rows = rows, warnings = requirementWarnings(build.displayStats, after) }
	return result
end

return function(build, request)
  local tab = build.itemsTab
  local item = request.raw and new("Item", request.raw) or tab.items[request.item_id]
  if not item or not item.base then return nil end
  local slot = request.slot and tab.slots[request.slot]
  if request.slot and (not slot or slot.nodeId) then return nil end
  local result = inspectItem(build, item, slot)
  if request.compare and result.compatible then return compareItem(build, item, slot, result) end
  return result
end
