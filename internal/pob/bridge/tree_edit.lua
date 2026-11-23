local function checkNode(node)
	assert(node and (node.type == "Normal" or node.type == "Notable"), "Выберите обычный пассив или значимое умение")
	assert(not node.ascendancyName and not node.containJewelSocket and not node.isMultipleChoice
		and not node.isMultipleChoiceOption and not node.isGrantedPassive and not node.isFreeAllocate,
		"Путь затрагивает особый пассив")
	assert(node.allocMode == 0, "Путь затрагивает ветвь комплекта оружия")
end

local function checkAllocation(build, node)
	checkNode(node)
	assert(not node.unlockConstraint or build.treeTab.viewer:checkUnlockConstraints(build, node),
		"Условия доступа к узлу не выполнены")
end

local function allocate(build, node)
	checkAllocation(build, node)
	local spec = build.spec
	assert(not node.alloc, "Пассив уже назначен")
	local path = spec:GetEffectiveAllocationPath(node)
	assert(path and #path > 0, "Нет доступного пути к узлу")
	if #node.intuitiveLeapLikesAffecting == 0 then
		for _, step in ipairs(path) do
			checkAllocation(build, step)
		end
	end
	spec:AllocNode(node)
	assert(node.alloc, "PoB не назначил выбранный пассив")
end

local function refund(build, node)
	assert(node.alloc, "Пассив ещё не назначен")
	for _, step in ipairs(node.depends) do
		checkNode(step)
	end
	build.spec:DeallocNode(node)
	assert(not node.alloc, "PoB не снял выбранный пассив")
end

local attributeNames = { Strength = "strength", Dexterity = "dexterity", Intelligence = "intelligence" }

local function addNode(change, spec, node)
	checkNode(node)
	change.added[#change.added + 1] = node.id
	if node.isAttribute then
		local override = assert(spec.hashOverrides[node.id], "PoB не выбрал атрибут узла")
		local list = change[assert(attributeNames[override.dn], "Неизвестный атрибут узла")]
		list[#list + 1] = node.id
	end
end

local function collectChange(build, before)
	local spec = build.spec
	local change = { added = {}, removed = {}, strength = {}, dexterity = {}, intelligence = {} }
	for id, node in pairs(spec.allocNodes) do
		if before[id] == nil then
			addNode(change, spec, node)
		else
			assert(node.allocMode == before[id], "Правка меняет ветвь комплекта оружия")
		end
	end
	for id in pairs(before) do
		if not spec.allocNodes[id] then
			checkNode(spec.nodes[id])
			change.removed[#change.removed + 1] = id
		end
	end
	for _, ids in pairs(change) do table.sort(ids) end
	return change
end

return function(build, request)
	local spec = build.spec
	assert(build.treeTab.activeSpec == request.tree_spec and spec.treeVersion == request.tree_version,
		"PoB загрузил другую спецификацию или версию дерева")
	local node = spec.nodes[request.node_id]
	checkNode(node)
	local before = {}
	for id, allocated in pairs(spec.allocNodes) do before[id] = allocated.allocMode end
	spec.allocMode = 0
	if request.action == "allocate" then
		allocate(build, node)
	else
		assert(request.action == "refund", "Неизвестное действие дерева")
		refund(build, node)
	end
	return collectChange(build, before)
end
