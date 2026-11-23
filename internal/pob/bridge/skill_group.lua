-- Правка групп повторяет редактор pob-poe2-mcp 312b337fa4a24f5ef55ebd38c8fd7bd46c7de78a (MIT).
local directory = assert(os.getenv("POB2_BRIDGE"))
local groupMethods = dofile(directory .. "/lua/gem_groups.lua")
local granted = dofile(directory .. "/lua/granted_supports.lua")

local function fail(_, message)
	error(message, 0)
end

local function finish()
	build.buildFlag = true
	runCallback("OnFrame")
	return {}
end

local function addGroup(label, gems)
	local node = { elem = "Skill", attrib = { enabled = "true", includeInFullDPS = "true", label = label or "", mainActiveSkill = "1" } }
	for _, input in ipairs(gems) do
		local gem = assert(build.data.gems[input.id], "Неизвестный ID камня")
		node[#node + 1] = { elem = "Gem", attrib = { skillId = gem.grantedEffectId, nameSpec = gem.name,
			gemId = gem.gameId, variantId = gem.variantId, level = tostring(input.level),
			quality = tostring(input.quality), enabled = tostring(input.enabled ~= false) } }
	end
	local skills = build.skillsTab
	skills:LoadSkill(node, skills.activeSkillSetId)
	skills:SetDisplayGroup(skills.socketGroupList[#skills.socketGroupList])
	finish()
end

local function replacementInput(group)
	local gems = {}
	for _, input in ipairs(group.gems or {}) do
		gems[#gems + 1] = { id = input.id, level = input.level, quality = input.quality,
			enabled = input.enabled, sourceIndex = input.source_index }
	end
	return { groupId = group.group_id, label = group.label, gems = gems }
end

local methods = {}
groupMethods(methods, fail, finish, granted)
return function(edit)
	local params = replacementInput(edit)
	if edit.delete then
		methods.deleteSkillGroup(params)
	elseif edit.group_id then
		methods.updateSkillGroup(params)
	else
		addGroup(params.label, params.gems)
	end
	return assert(build:SaveDB("Редактор групп"))
end
