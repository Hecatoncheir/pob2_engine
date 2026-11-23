-- Выбор повторяет штатные списки Lua PoB. Возвращает адрес поля в XML.
local names = { groups = "mainSocketGroup", skills = "mainSkill", parts = "mainSkillPart", stat_sets = "statSet" }

local function sourceGem(group)
	local skill = group.displaySkillList[group.mainActiveSkill]
	local source = skill and skill.activeEffect.srcInstance
	for index, gem in ipairs(group.gemList) do
		if gem == source then return index end
	end
	error("PoB не нашёл исходный камень выбранного умения")
end

return function(build, edit, validate)
	local control = build.controls[names[edit.control]]
	local option = control and control.list[edit.option_id]
	assert(option and control:IsShown() and control:IsEnabled(), "Выбор умения недоступен")
	local gemOption = edit.control == "parts" or edit.control == "stat_sets"
	assert(not gemOption or #control.list > 1, "Выбор варианта недоступен")
	if not validate then
		control:SetSel(control:ListIndexToDropIndex(edit.option_id, 0))
		runCallback("OnFrame")
	end
	assert(control.selIndex == edit.option_id, "PoB не подтвердил выбранный вариант")
	local index = build.mainSocketGroup
	local group = build.skillsTab.socketGroupList[index]
	assert(group, "PoB не нашёл выбранную группу")
	return { group = index, active_skill = group.mainActiveSkill,
		gem = gemOption and sourceGem(group) or nil,
		effect = edit.control == "stat_sets" and option.grantedEffectId or nil }
end
