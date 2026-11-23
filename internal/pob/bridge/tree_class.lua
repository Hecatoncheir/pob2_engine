-- Варианты соединения и сброса повторяют treeOptions из pob-poe2-mcp
-- (312b337fa4a24f5ef55ebd38c8fd7bd46c7de78a, MIT).
local function selectClass(spec, edit)
	if edit.class_id == spec.curClassId then
		if edit.mode == "reset" then
			wipeTable(spec.hashOverrides)
			spec:ResetNodes()
		end
		return
	end
	if spec:CountAllocNodes() > 0 and not spec:IsClassConnected(edit.class_id) then
		if edit.mode == "connect" then
			assert(spec:ConnectToClass(edit.class_id), "Не удалось соединить класс")
		elseif edit.mode ~= "reset" then
			error("Смена класса требует соединения или сброса дерева", 0)
		end
	end
	if edit.mode == "reset" then
		wipeTable(spec.hashOverrides)
		spec:ResetNodes()
	end
	spec:SelectClass(edit.class_id)
end

local function savedClass(spec)
	local saved = {}
	spec:Save(saved)
	local weapons = { {}, {} }
	for id, node in pairs(spec.allocNodes) do
		local weapon = weapons[node.allocMode]
		if weapon then weapon[#weapon + 1] = id end
	end
	table.sort(weapons[1])
	table.sort(weapons[2])
	local result = {
		identity = { class = spec.curClassName, ascendancy = spec.curAscendClassName or "None" },
		spec = saved.attrib,
		weapon_one = table.concat(weapons[1], ","),
		weapon_two = table.concat(weapons[2], ","),
	}
	for _, child in ipairs(saved) do
		if child.elem == "URL" then result.url = child[1] end
		if child.elem == "Overrides" and child[1] then result.attributes = child[1].attrib end
	end
	return result
end

return function(build, request)
	local spec, edit = build.spec, request.tree_class
	assert(build.treeTab.activeSpec == request.tree_spec and (not request.tree_version or spec.treeVersion == request.tree_version),
		"PoB загрузил другую спецификацию или версию дерева")
	local class = assert(spec.tree.classes[edit.class_id], "Класс не найден")
	assert(class.classes[edit.ascendancy_id], "Восхождение недоступно для класса")
	selectClass(spec, edit)
	if spec.curAscendClassId ~= edit.ascendancy_id then
		spec:SelectAscendClass(edit.ascendancy_id)
	end
	spec:BuildAllDependsAndPaths()
	return savedClass(spec)
end
