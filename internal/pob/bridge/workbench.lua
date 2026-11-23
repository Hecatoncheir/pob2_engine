-- Запускает совместимые инструменты мастерской на той же установке PoB.
POB_MCP_PACKAGE_DIR = assert(os.getenv("POB2_BRIDGE"))
POB_MCP_RUNTIME_DIR = assert(os.getenv("POB2_RUNTIME"))
POB_MCP_HEADLESS = true
local native = assert(os.getenv("POB2_NATIVE"))
local extension = assert(os.getenv("POB2_LIBRARY_EXTENSION"))
package.path = POB_MCP_RUNTIME_DIR .. "/lua/?.lua;" .. POB_MCP_RUNTIME_DIR .. "/lua/?/init.lua;./?.lua;./?/init.lua"
package.cpath = native .. "/?." .. extension .. ";" .. POB_MCP_RUNTIME_DIR .. "/?." .. extension
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
