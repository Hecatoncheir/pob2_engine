-- Числа, порядок и подписи вычисляет штатный отчёт Lua PoB.
local function metricID(stat)
	return stat.stat or (stat.combinedOffDef and "OffenceDefence")
end

local function metrics(treeTab)
	local entries = {}
	for _, stat in ipairs(treeTab.powerStatList) do
		local id = metricID(stat)
		if id and not stat.ignoreForNodes then
			entries[#entries + 1] = { id = id, label = stat.label }
		end
	end
	return entries
end

local function calculate(build, stat, distance)
	build.spec:BuildAllDependsAndPaths()
	local tab = build.calcsTab
	tab.powerStat, tab.nodePowerMaxDepth, tab.powerMax = stat, distance, nil
	local thread = coroutine.create(tab.PowerBuilder)
	repeat
		local ok, message = coroutine.resume(thread, tab)
		assert(ok, message)
	until coroutine.status(thread) == "dead"
end

local function filterRows(build, stat, distance, allocation)
	local result = {}
	for _, row in ipairs(build.treeTab:BuildPowerReportList(stat)) do
		local allocated = row.allocated == true
		local visible = allocation == "all" or (allocation == "allocated") == allocated
		local inRange = type(row.pathDist) == "number" and (not distance or row.pathDist <= distance)
		if visible and inRange and row.power ~= 0 and build.spec.nodes[row.id] then
			result[#result + 1] = { id = row.id, name = row.name, kind = row.type, allocated = allocated,
				distance = row.pathDist, power = row.power, path_power = row.pathPower,
				power_text = row.powerStr, path_power_text = row.pathPowerStr }
		end
	end
	return result
end

return function(build, request)
	if not request.metric then return { metrics = metrics(build.treeTab) } end
	for _, stat in ipairs(build.treeTab.powerStatList) do
		if metricID(stat) == request.metric and not stat.ignoreForNodes then
			calculate(build, stat, request.max_distance)
			return { metric = { id = request.metric, label = stat.label },
				entries = filterRows(build, stat, request.max_distance, request.allocation) }
		end
	end
	return nil
end
