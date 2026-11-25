-- Пакетный поиск повторяет порядок узлов в PassiveSpec из PoB v0.23.1.
local minGoRoots, maxWords = 5, 2^24
local ok, ffi = pcall(require, "ffi")
if not ok then return function() end end

local hasTableNew, tableNew = pcall(require, "table.new")
if not hasTableNew or type(tableNew) ~= "function" then
	tableNew = function() return {} end
end

local function unlocked(spec, node)
	if not node.unlockConstraint then return true end
	for _, id in ipairs(node.unlockConstraint.nodes) do
		if not assert(spec.nodes[id]).alloc then return false end
	end
	return true
end

local function flags(spec, node)
	return (node.alloc and 1 or 0) + (unlocked(spec, node) and 2 or 0)
		+ (node.type == "Mastery" and 4 or 0)
		+ ((node.type == "ClassStart" or node.type == "AscendClassStart") and 8 or 0)
end

local function reserve(buffer, words)
	assert(words <= maxWords, "переполнен буфер дерева")
	if words > buffer.capacity then
		local capacity = math.min(maxWords, math.max(words, buffer.capacity * 2))
		buffer.data, buffer.capacity = ffi.new("uint32_t[?]", capacity), capacity
	end
	return buffer.data
end

local function collectNodes(spec)
	local nodes, indices, edges = {}, {}, 0
	for _, node in pairs(spec.nodes) do
		indices[node] = #nodes
		nodes[#nodes + 1] = node
		edges = edges + #node.linked
	end
	return nodes, indices, edges
end

local function pack(spec, roots, inputBuffer)
	local nodes, indices, edges = collectNodes(spec)
	local field, edge = 3, 3 + #nodes * 5
	local rootStart = edge + edges
	local words = rootStart + #roots
	local input = reserve(inputBuffer, words)
	input[0], input[1], input[2] = #nodes, edges, #roots
	local ascendancies = {}
	for index, node in ipairs(nodes) do
		local name, ascendancy = node.ascendancyName, 0
		if name ~= nil then
			assert(type(name) == "string", "неизвестный тип имени восхождения")
			ascendancy = ascendancies[name] or index
			ascendancies[name] = ascendancy
		end
		assert(node.pathDist == 0 or node.pathDist == 1000, "исходные пути уже обновлены")
		input[field], input[field + 1], input[field + 2] = ascendancy, node.allocMode or 0, node.pathDist
		input[field + 3], input[field + 4] = flags(spec, node), #node.linked
		field = field + 5
		for _, neighbor in ipairs(node.linked) do
			input[edge] = assert(indices[neighbor], "сосед не входит в дерево")
			edge = edge + 1
		end
	end
	for index, root in ipairs(roots) do
		input[rootStart + index - 1] = assert(indices[root], "корень не входит в дерево")
	end
	return nodes, edges, input, words
end

local function restorePath(nodes, output, offset, distance, tail)
	local path = tableNew(math.min(distance, 128), 0)
	local index = 1
	while tail > 0 do
		local link = offset + tail * 2
		path[index] = nodes[output[link]]
		index, tail = index + 1, output[link + 1]
	end
	return path
end

local function restorePaths(nodes, output)
	local offset = 1 + #nodes * 3 - 2
	for index, node in ipairs(nodes) do
		local field = index * 3 - 2
		local distance, root, tail = output[field], output[field + 1], output[field + 2]
		local path = root > 0 and restorePath(nodes, output, offset, distance, tail) or nil
		node.pathDist, node.pathRoot, node.path = distance, nodes[root], path
	end
end

local function solve(spec, roots, inputBuffer, outputBuffer, callback)
	local nodes, edges, input, words = pack(spec, roots, inputBuffer)
	local output = reserve(outputBuffer, 1 + #nodes * 64 + edges * 2)
	local written = callback(input, words, output, outputBuffer.capacity)
	local expected = 1 + #nodes * 3 + output[0] * 2
	if written <= outputBuffer.capacity and written == expected then
		local restored = pcall(restorePaths, nodes, output)
		if restored then return true end
		for index, node in ipairs(nodes) do
			node.pathDist, node.pathRoot, node.path = input[index * 5], nil, nil
		end
	end
	return false
end

local function collectRoots(spec, rebuild)
	local roots = {}
	local previous = rawget(spec, "BuildPathFromNode")
	spec.BuildPathFromNode = function(_, root) roots[#roots + 1] = root end
	local rebuilt, failure = pcall(rebuild, spec)
	spec.BuildPathFromNode = previous
	if not rebuilt then error(failure, 0) end
	return roots
end

return function(class, address)
	if not class then return end
	local callback = ffi.cast("uint32_t(*)(const uint32_t*,uint32_t,uint32_t*,uint32_t)", address)
	local rebuild, search = class.BuildAllDependsAndPaths, class.BuildPathFromNode
	local canPath, jewelDistance = class.CanPathThroughAllocMode, class.SetNodeDistanceToClassStart
	local input, output = { capacity = 0 }, { capacity = 0 }

	class.BuildAllDependsAndPaths = function(spec)
		if spec.BuildPathFromNode ~= search or spec.CanPathThroughAllocMode ~= canPath
			or spec.SetNodeDistanceToClassStart ~= jewelDistance then
			return rebuild(spec)
		end
		local roots = collectRoots(spec, rebuild)
		if #roots >= minGoRoots then
			local solved, restored = pcall(solve, spec, roots, input, output, callback)
			if solved and restored then return end
		end
		for _, root in ipairs(roots) do search(spec, root) end
	end
end
