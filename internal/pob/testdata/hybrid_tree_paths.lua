loadBuildFromXML(xml, "Witch")

local class = common.classes.PassiveSpec
local rebuild, search = class.BuildAllDependsAndPaths, class.BuildPathFromNode
local luaSearches = 0
class.BuildPathFromNode = function(spec, root)
	luaSearches = luaSearches + 1
	return search(spec, root)
end

local function snapshot(spec)
	local saved = {}
	for id, node in pairs(spec.nodes) do
		saved[id] = { node.pathDist, node.pathRoot, node.path, node.distanceToClassStart }
	end
	return saved
end

local function comparePath(node, saved)
	assert(node.pathDist == saved[1] and node.pathRoot == saved[2] and node.distanceToClassStart == saved[4],
		"изменились расстояние или корень узла " .. node.id)
	local got, expected = node.path, saved[3]
	assert((got == nil) == (expected == nil), "изменилось наличие пути узла " .. node.id)
	if not got then return end
	assert(#got == #expected, "изменилась длина пути узла " .. node.id)
	for index, step in ipairs(got) do
		assert(step == expected[index], "изменился порядок пути узла " .. node.id)
	end
end

local function compare(spec, expectMany, expectGo)
	luaSearches = 0
	rebuild(spec)
	local roots = luaSearches
	assert(roots > 0 and (roots >= 5) == expectMany, "неподходящее число корней " .. roots)
	local expected = snapshot(spec)
	luaSearches = 0
	spec:BuildAllDependsAndPaths()
	assert(luaSearches == (expectGo and 0 or roots),
		string.format("выбран неверный алгоритм: Go=%s, Lua=%d, корней=%d", tostring(expectGo), luaSearches, roots))
	for id, node in pairs(spec.nodes) do comparePath(node, expected[id]) end
end

local install = dofile(directory .. "/tree_paths.lua")
assert(POB2_GO_PATH_CALLBACK, "движок не зарегистрировал Go-обход")
install(class, POB2_GO_PATH_CALLBACK)
compare(build.spec, false, false)

local edit = dofile(directory .. "/tree_edit.lua")
edit(build, {
	tree_spec = build.treeTab.activeSpec, tree_version = build.spec.treeVersion,
	node_id = 51184, action = "allocate", allocation_mode = "weapon_set_1",
})
compare(build.spec, true, true)

-- Нераспознанное значение режима отклоняет пакет, но Lua продолжает обход.
local untouched = build.spec.nodes[13397]
assert(not untouched.alloc, "неподходящий узел для проверки резерва")
untouched.allocMode = 3
compare(build.spec, true, false)
