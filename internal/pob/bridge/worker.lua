-- Выполняет только расчёт переданного снимка. Сессия принадлежит Go.
local directory = assert(os.getenv("POB2_BRIDGE"))
local runtime = assert(os.getenv("POB2_RUNTIME"))
local native = assert(os.getenv("POB2_NATIVE"))
local extension = assert(os.getenv("POB2_LIBRARY_EXTENSION"))
POB_MCP_PACKAGE_DIR = directory
POB_MCP_RUNTIME_DIR = runtime
POB_MCP_HEADLESS = true
package.path = runtime .. "/lua/?.lua;" .. runtime .. "/lua/?/init.lua;./?.lua;./?/init.lua"
package.cpath = native .. "/?." .. extension .. ";" .. runtime .. "/?." .. extension
arg = {}

local nativeLoadfile = loadfile
function loadfile(path)
	local file, err = io.open(path, "rb")
	if not file then return nil, err end
	local first = file:read(1)
	file:close()
	if first ~= "#" then return nativeLoadfile(path) end
	local sourceFile = assert(io.open(path, "rb"))
	local source = sourceFile:read("*a")
	sourceFile:close()
	return loadstring(source:gsub("^[^\n]*", "", 1), "@" .. path)
end
function dofile(path)
	return assert(loadfile(path))()
end

local json = require("dkjson")
dofile(directory .. "/lua/headless_bootstrap.lua")
local catalog = dofile(directory .. "/catalog.lua")
local diagnostics = dofile(directory .. "/diagnostics.lua")
local editTree = dofile(directory .. "/tree_edit.lua")
local editTreeClass = dofile(directory .. "/tree_class.lua")
local readTreeGeometry = dofile(directory .. "/tree_geometry.lua")
local readTreeNodeTooltip = dofile(directory .. "/tree_node_tooltip.lua")
local readBuildReport = dofile(directory .. "/build_report.lua")
local selectSkillChoice = dofile(directory .. "/skill_choice.lua")
local editSkillGroup = dofile(directory .. "/skill_group.lua")
local readTreePower = dofile(directory .. "/tree_power.lua")
local drawTreeScene = dofile(directory .. "/tree_scene.lua")
local readEquipmentDetails = dofile(directory .. "/equipment_details.lua")

local function clean(text)
	return tostring(text or ""):gsub("%^x%x%x%x%x%x%x", ""):gsub("%^%d", "")
end

local function checkModifiers(text, unsupported)
	for line in (text or ""):gmatch("[^\r\n]+") do
		if line:match("%S") then
			local mods, extra = modLib.parseMod(line)
			if not mods or extra then
				unsupported[#unsupported + 1] = clean(line)
			end
		end
	end
end

local function unsupportedModifiers()
	local unsupported = {}
	local config = build.configTab
	checkModifiers(config.input.customMods, unsupported)
	local active = config.configSets[config.activeConfigSetId]
	for _, block in ipairs(active and active.customModsList or {}) do
		if block.enabled then checkModifiers(block.text, unsupported) end
	end
	return unsupported
end

local function loadDocument(xml, name)
	local parsed = common.xml.ParseXML(xml)
	assert(parsed and parsed[1] and parsed[1].elem == "PathOfBuilding2", "Ожидался XML PathOfBuilding2")
	loadBuildFromXML(xml, name or "PoB2Engine")
	assert(not launch.promptMsg, tostring(launch.promptMsg))
end

local function validateEquipment(request)
	loadDocument(assert(request.xml))
	local tab, name, id = build.itemsTab, request.slot, request.item_id or 0
	local slot = tab.slots[name]
	if request.node_id then
		local node = build.spec.nodes[request.node_id]
		if not node or node.type ~= "Socket" or node.ascendancyName or node.expansionJewel then
			return { error = "Выберите обычное гнездо самоцвета" }
		end
		if not node.alloc then
			return { error = "Сначала назначьте гнездо самоцвета" }
		end
		slot = tab:GetSocketAndJewelForNodeID(node.id)
	elseif not tab.activeItemSet[name] then
		slot = nil
	end
	if not slot then
		return { error = "PoB не распознал слот экипировки " .. name }
	end
	local item = tab.items[id]
	if id ~= 0 and (not item or not item.base) then
		return { error = "PoB не распознал базу предмета" }
	end
	if id ~= 0 and not tab:IsItemValidForSlot(item, name) then
		return { error = "Предмет несовместим со слотом " .. name }
	end
	if (slot.selItemId or 0) ~= id then
		return { error = "PoB не применил правку слота " .. name }
	end
	return { equipment_valid = true }
end

local function calculate(xml)
	loadDocument(xml)
	local output = assert(build.calcsTab.mainOutput)
	local stats, warnings = {}, {}
	for key, value in pairs(output) do
		if type(key) == "string" and type(value) == "number" and value == value and math.abs(value) < math.huge then
			stats[key] = value
		end
	end
	for _, warning in ipairs(build.controls.warnings.lines or {}) do
		warnings[#warnings + 1] = clean(warning)
	end
	local skills = build.controls.mainSkill
	local selected = skills and skills.list and skills.list[skills.selIndex]
	local skill = type(selected) == "table" and selected.label or selected
	return { stats = stats, warnings = warnings, unsupported = unsupportedModifiers(), issues = diagnostics.collect(), skill = clean(skill), mode = "EFFECTIVE" }
end

local function selectTreeSpec(request)
	loadDocument(assert(request.xml))
	local spec = build.spec
	assert(build.treeTab.activeSpec == request.tree_spec and (not request.tree_version or spec.treeVersion == request.tree_version),
		"PoB загрузил другую спецификацию или версию дерева")
	return { tree_identity = { class = spec.curClassName, ascendancy = spec.curAscendClassName or "None" } }
end

local function reply(value)
	io.stdout:write(json.encode(value), "\n")
	io.stdout:flush()
end

-- Дельта штатной отрисовки по образцу image_data.lua из pob-poe2-mcp (MIT).
local baseSceneXML, baseSceneKeys
local function drawCommandKey(command)
	return json.encode({command.file, command.index, command.positions, command.uv,
		command.color, command.blend, command.clamp, command.mask, command.layer, command.subLayer})
end

local function cacheBaseSceneKeys(xml)
	local keys = {}
	for _, command in ipairs(drawTreeScene(build, false).commands) do keys[drawCommandKey(command)] = true end
	baseSceneXML, baseSceneKeys = xml, keys
end

local function hoverTreeScene(xml, nodeID)
	if baseSceneXML ~= xml then cacheBaseSceneKeys(xml) end
	local scene = drawTreeScene(build, false, nodeID)
	local overlay = {}
	for _, command in ipairs(scene.commands) do
		if not baseSceneKeys[drawCommandKey(command)] then overlay[#overlay + 1] = command end
	end
	scene.commands = overlay
	return scene
end

reply({ ready = true })
for line in io.lines() do
	local ok, result = pcall(function()
		local request = assert(json.decode(line))
		if request.operation == "catalog" then return { catalog = catalog.read() } end
		if request.operation == "validate_equipment" then return validateEquipment(request) end
		if request.operation == "select_tree_spec" then return selectTreeSpec(request) end
		if request.operation == "edit_tree" then
			loadDocument(assert(request.xml))
			return { tree_edit = editTree(build, request) }
		end
		if request.operation == "edit_tree_class" then
			loadDocument(assert(request.xml))
			return { tree_class = editTreeClass(build, request) }
		end
		if request.operation == "tree_geometry" then
			loadDocument(assert(request.xml))
			local geometry = readTreeGeometry(build.spec)
			if request.include_scene then geometry.scene = drawTreeScene(build, false) end
			return { tree_geometry = geometry }
		end
		if request.operation == "tree_node_tooltip" then
			loadDocument(assert(request.xml))
			local tooltip = readTreeNodeTooltip(build, request)
			if tooltip and request.include_scene then tooltip.scene = hoverTreeScene(request.xml, tooltip.node_id) end
			return { tree_tooltip = tooltip or json.null }
		end
		if request.operation == "build_report" then
			loadDocument(assert(request.xml), request.build_name)
			return { build_report = readBuildReport(build, clean) }
		end
		if request.operation == "skill_choice" then
			loadDocument(assert(request.xml))
			return { skill_choice = selectSkillChoice(build, request.skill_choice, request.validate) }
		end
		if request.operation == "skill_group" then
			loadDocument(assert(request.xml))
			return { skill_group_xml = editSkillGroup(request.skill_group) }
		end
		if request.operation == "tree_power" then
			loadDocument(assert(request.xml))
			if request.report_progress then
				GetTime = function() return os.clock() * 1000 end
				build.powerBuilderProgressCallback = function(percent)
					reply({ tree_power_progress = percent })
				end
			end
			local power = readTreePower(build, request.tree_power)
			if power and power.metric and request.tree_power.include_scene then power.scene = drawTreeScene(build, true) end
			return { tree_power = power or json.null }
		end
		if request.operation == "equipment_details" then
			loadDocument(assert(request.xml))
			return { equipment_details = readEquipmentDetails(build, request.equipment_details) or json.null }
		end
		return calculate(assert(request.xml))
	end)
	if ok then reply(result) else reply({ error = clean(result) }) end
end
