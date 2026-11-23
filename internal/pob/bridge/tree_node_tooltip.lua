-- Извлекает строки из штатной подсказки PoB и сохраняет её коды цвета.
local function appendLines(result, tooltip)
	for _, line in ipairs(tooltip.lines) do
		result[#result + 1] = { size = line.size, text = line.text, separator = line.text == nil }
	end
end

local function increasedEffect(spec)
	local effect = 0
	for _, node in pairs(spec.allocNodes) do
		effect = effect + node.modList:Sum("INC", nil, "SmallPassiveSkillEffect")
	end
	return effect
end

return function(build, request)
	local spec = build.spec
	local node = spec.nodes[request.node_id]
	if not node or node.type == "OnlyImage" then return nil end
	spec:BuildAllDependsAndPaths()
	local viewer = new("PassiveTreeView")
	viewer.showStatDifferences = request.compare == true
	local tooltip = new("Tooltip")
	viewer:AddNodeTooltip(tooltip, node, build, increasedEffect(spec))
	local lines, path, recipe = {}, {}, {}
	appendLines(lines, tooltip)
	if #viewer.skillTooltip.lines > 0 then
		lines[#lines + 1] = { separator = true }
		appendLines(lines, viewer.skillTooltip)
	end
	for _, member in ipairs(node.alloc and node.depends or node.path or {}) do
		path[#path + 1] = member.id
	end
	for _, ingredient in ipairs(tooltip.recipe or {}) do
		recipe[#recipe + 1] = ingredient.name
	end
	local points = node.pathDist
	if points and (points < 0 or points >= math.huge) then points = nil end
	return { node_id = node.id, lines = lines, path = path, points = points,
		note = spec.nodeNotes and spec.nodeNotes[node.id], recipe = recipe }
end
