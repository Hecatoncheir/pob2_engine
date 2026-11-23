local function checkNode(node, rules)
	assert(node and (node.type == "Normal" or node.type == "Notable" or node == rules.socket),
		"Выберите обычный пассив, значимый узел или пустое гнездо самоцвета")
	assert(not node.ascendancyName and not node.containJewelSocket and not node.isMultipleChoice
		and not node.isMultipleChoiceOption and not node.isGrantedPassive and not node.isFreeAllocate,
		"Путь затрагивает особый пассив")
	assert(rules.usesModes or node.allocMode == 0, "Путь затрагивает ветвь комплекта оружия")
end

local function checkSocket(spec, node, mode)
	if not node or node.type ~= "Socket" then return end
	assert(not mode or mode == "common", "Гнездо назначается в общем режиме")
	assert(not spec.jewels[node.id] or spec.jewels[node.id] == 0, "Сначала выньте самоцвет из гнезда")
	return node
end

local function checkAllocation(build, node, rules)
	checkNode(node, rules)
	assert(not node.unlockConstraint or build.treeTab.viewer:checkUnlockConstraints(build, node),
		"Условия доступа к узлу не выполнены")
end

local function allocate(build, node, attribute, rules)
	checkAllocation(build, node, rules)
	local spec = build.spec
	assert(not node.alloc, "Пассив уже назначен")
	local path = spec:GetEffectiveAllocationPath(node)
	assert(path and #path > 0, "Нет доступного пути к узлу")
	if #node.intuitiveLeapLikesAffecting == 0 then
		for _, step in ipairs(path) do
			checkAllocation(build, step, rules)
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

local function refund(build, node, rules)
	assert(node.alloc, "Пассив ещё не назначен")
	for _, step in ipairs(node.depends) do
		checkNode(step, rules)
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

local allocationNames = { [0] = "common", [1] = "weapon_set_1", [2] = "weapon_set_2" }
local allocationModes = { common = 0, weapon_set_1 = 1, weapon_set_2 = 2 }

local function recordAllocation(change, id, before, after)
	if not change.allocations or before == after then return end
	change.allocations[#change.allocations + 1] = {
		node_id = id, before = allocationNames[before], after = allocationNames[after],
	}
end

local function collectRemoved(change, spec, before, rules)
	for id, mode in pairs(before) do
		if not spec.allocNodes[id] then
			checkNode(spec.nodes[id], rules)
			change.removed[#change.removed + 1] = id
			recordAllocation(change, id, mode, nil)
		end
	end
end

local function collectChange(spec, before, rules)
	local change = { added = {}, removed = {}, allocations = rules.usesModes and {} or nil }
	for id, node in pairs(spec.allocNodes) do
		if before[id] == nil then
			checkNode(node, rules)
			change.added[#change.added + 1] = id
		else
			assert(rules.usesModes or node.allocMode == before[id], "Правка меняет ветвь комплекта оружия")
		end
		recordAllocation(change, id, before[id], node.allocMode)
	end
	collectRemoved(change, spec, before, rules)
	table.sort(change.added)
	table.sort(change.removed)
	if change.allocations then
		table.sort(change.allocations, function(left, right) return left.node_id < right.node_id end)
	end
	change.attributes = allocatedAttributes(spec, change.added)
	return change
end

local function pointUsage(spec)
	local total, _, _, _, weapon1, weapon2 = spec:CountAllocNodes()
	return { passives = total - math.min(weapon1, weapon2), weapon_set_1 = weapon1, weapon_set_2 = weapon2 }
end

local function applyEdit(build, node, request, rules)
	local spec = build.spec
	local index = request.attribute and assert(attributeIndices[request.attribute], "Неизвестный атрибут узла")
	if request.action == "allocate" then
		allocate(build, node, index, rules)
		assert(node.allocMode == spec.allocMode, "PoB назначил пассив в другом режиме")
		return
	end
	assert(node.allocMode == spec.allocMode, "Пассив назначен для другого режима дерева")
	if request.action == "set_attribute" then
		return setAttribute(spec, node, request.attribute, assert(index))
	end
	assert(request.action == "refund", "Неизвестное действие дерева")
	refund(build, node, rules)
end

return function(build, request)
	local spec = build.spec
	assert(build.treeTab.activeSpec == request.tree_spec and spec.treeVersion == request.tree_version,
		"PoB загрузил другую спецификацию или версию дерева")
	local node = spec.nodes[request.node_id]
	local usesModes = request.allocation_mode ~= nil
	local rules = { usesModes = usesModes, socket = checkSocket(spec, node, request.allocation_mode) }
	checkNode(node, rules)
	local before = {}
	for id, allocated in pairs(spec.allocNodes) do before[id] = allocated.allocMode end
	local beforePoints = usesModes and pointUsage(spec)
	spec.allocMode = usesModes and assert(allocationModes[request.allocation_mode], "Неизвестный режим дерева") or 0
	local attributeChange = applyEdit(build, node, request, rules)
	local change = collectChange(spec, before, rules)
	if attributeChange then change.attributes[#change.attributes + 1] = attributeChange end
	if usesModes then change.points = { before = beforePoints, after = pointUsage(spec) } end
	return change
end
