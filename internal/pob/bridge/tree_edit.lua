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

local function allocate(build, node, attribute)
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
	if attribute then
		assert(node.isAttribute, "Выберите узел с выбором атрибута")
		spec:SwitchAttributeNode(node.id, attribute)
		spec.attributeIndex = attribute
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
local attributeIndices = { strength = 1, dexterity = 2, intelligence = 3 }

local function chosenAttribute(spec, node)
	return attributeNames[(spec.hashOverrides[node.id] or node).dn]
end

local function setAttribute(spec, node, attribute, index)
	assert(node.isAttribute, "Выберите узел с выбором атрибута")
	assert(node.alloc, "Сначала назначьте пассив")
	local before = chosenAttribute(spec, node)
	assert(before ~= attribute, "Этот атрибут уже выбран")
	spec:SwitchAttributeNode(node.id, index)
	spec:BuildAllDependsAndPaths()
	local after = chosenAttribute(spec, spec.nodes[node.id])
	assert(after == attribute, "PoB не подтвердил новый атрибут узла")
	return { node_id = node.id, before = before, after = after }
end

local function allocatedAttributes(spec, nodes)
	local attributes = {}
	for _, id in ipairs(nodes) do
		local node = spec.allocNodes[id]
		if node.isAttribute then
			local after = assert(chosenAttribute(spec, node), "PoB не выбрал атрибут узла")
			attributes[#attributes + 1] = { node_id = id, after = after }
		end
	end
	return attributes
end

local function collectChange(spec, before)
	local change = { added = {}, removed = {} }
	for id, node in pairs(spec.allocNodes) do
		if before[id] == nil then
			checkNode(node)
			change.added[#change.added + 1] = id
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
	table.sort(change.added)
	table.sort(change.removed)
	change.attributes = allocatedAttributes(spec, change.added)
	return change
end

return function(build, request)
	local spec = build.spec
	assert(build.treeTab.activeSpec == request.tree_spec and spec.treeVersion == request.tree_version,
		"PoB загрузил другую спецификацию или версию дерева")
	local node = spec.nodes[request.node_id]
	checkNode(node)
	local index = request.attribute and assert(attributeIndices[request.attribute], "Неизвестный атрибут узла")
	local before = {}
	for id, allocated in pairs(spec.allocNodes) do before[id] = allocated.allocMode end
	spec.allocMode = 0
	local attributeChange
	if request.action == "allocate" then
		allocate(build, node, index)
	elseif request.action == "set_attribute" then
		attributeChange = setAttribute(spec, node, request.attribute, assert(index))
	else
		assert(request.action == "refund", "Неизвестное действие дерева")
		refund(build, node)
	end
	local change = collectChange(spec, before)
	if attributeChange then change.attributes[#change.attributes + 1] = attributeChange end
	return change
end
