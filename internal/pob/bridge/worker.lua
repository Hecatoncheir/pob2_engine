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

local function calculate(xml)
	local parsed = common.xml.ParseXML(xml)
	assert(parsed and parsed[1] and parsed[1].elem == "PathOfBuilding2", "Ожидался XML PathOfBuilding2")
	loadBuildFromXML(xml, "PoB2Engine")
	assert(not launch.promptMsg, tostring(launch.promptMsg))
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
	return { stats = stats, warnings = warnings, unsupported = unsupportedModifiers(), skill = clean(skill), mode = "EFFECTIVE" }
end

local function reply(value)
	io.stdout:write(json.encode(value), "\n")
	io.stdout:flush()
end

reply({ ready = true })
for line in io.lines() do
	local ok, result = pcall(function()
		local request = assert(json.decode(line))
		return calculate(assert(request.xml))
	end)
	if ok then reply(result) else reply({ error = clean(result) }) end
end
