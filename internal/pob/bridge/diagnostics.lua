local diagnostics = {}

local function gemProblem(gem, group)
	if not gem.enabled then return end
	if gem.errMsg then return "pob_gem_error", gem.errMsg end
	local granted = gem.gemData and gem.gemData.grantedEffect or gem.grantedEffect
	if not granted and gem.nameSpec and gem.nameSpec:match("%S") then
		return "pob_gem_error", "PoB не распознал камень " .. gem.nameSpec
	end
	return nil
end

local function supportProblem(gem, group)
	local effect = gem.supportEffect
	if not effect then return end
	if effect.superseded then
		return "pob_support_superseded", "Поддержка " .. gem.nameSpec .. " вытеснена другой поддержкой"
	end
	if not next(effect.isSupporting or {}) and #(group.displaySkillList or {}) > 0 then
		return "pob_support_incompatible", "Поддержка " .. gem.nameSpec .. " не применяется к активным умениям группы"
	end
end

local function groupProblems(result, id, group)
	if not group.enabled then return end
	for index, gem in ipairs(group.gemList or {}) do
		local code, message = gemProblem(gem, group)
		if not code and gem.enabled then code, message = supportProblem(gem, group) end
		if code then
			result[#result + 1] = { code = code, path = "Skills/Skill[" .. id .. "]/Gem[" .. index .. "]", message = message }
		end
	end
end

local function itemLines(result, item, slot)
	for _, name in ipairs({ "implicitModLines", "explicitModLines", "enchantModLines", "buffModLines" }) do
		for _, line in ipairs(item[name] or {}) do
			if line.extra or not line.modList or #line.modList == 0 then
				result[#result + 1] = { code = "unsupported_modifier", path = "Items/" .. slot, message = line.line or "Нераспознанное свойство предмета" }
			end
		end
	end
end

function diagnostics.collect()
	local result, seen = {}, {}
	for id, group in ipairs(build.skillsTab.socketGroupList or {}) do groupProblems(result, id, group) end
	for slot, item in pairs(build.calcsTab.mainEnv.player.itemList or {}) do
		if not seen[item] then
			seen[item] = true
			itemLines(result, item, slot)
		end
	end
	return result
end

return diagnostics
