-- Координаты и связи берутся из активной спецификации оригинального PoB,
-- включая подграфы, которые изменились от самоцветов.
local function hasPosition(node)
	return type(node.x) == "number" and type(node.y) == "number" and type(node.type) == "string" and node.type ~= "OnlyImage"
end

local function geometryNode(id, node)
	return { id = id, x = node.x, y = node.y, name = node.dn or node.name or "",
		kind = node.type, icon = node.icon, allocated = node.alloc == true,
		attribute = node.isAttribute == true, allocation_mode = node.allocMode or 0,
		draw_size = node.targetSize and node.targetSize.width or 0 }
end

local function nodeLinks(edges, id, node)
	local seen = {}
	for _, adjacent in ipairs(node.linked or {}) do
		local target = adjacent.id
		if id < target and not seen[target] and hasPosition(adjacent) then
			seen[target] = true
			edges[#edges + 1] = { from_node = id, to_node = target }
		end
	end
end

return function(spec)
	local nodes, edges = {}, {}
	for id, node in pairs(spec.nodes) do
		if hasPosition(node) then
			nodes[#nodes + 1] = geometryNode(id, node)
			nodeLinks(edges, id, node)
		end
	end
	table.sort(nodes, function(a, b) return a.id < b.id end)
	table.sort(edges, function(a, b)
		if a.from_node == b.from_node then return a.to_node < b.to_node end
		return a.from_node < b.from_node
	end)
	return { version = spec.treeVersion, nodes = nodes, edges = edges }
end
