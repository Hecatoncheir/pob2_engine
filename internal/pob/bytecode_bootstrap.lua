-- Если компилированный вход не загружается, выполняется исходный мост.
local root = assert(os.getenv("POB2_BRIDGE")) .. "/"
local compiled = root .. "compiled/"
local nativeLoadfile = loadfile
local main = nativeLoadfile(compiled .. entry, "b")
if not main then return false end

function loadfile(path, ...)
	if type(path) == "string" and path:sub(1, #root) == root then
		local chunk = nativeLoadfile(compiled .. path:sub(#root + 1), "b")
		if chunk then return chunk end
	end
	return nativeLoadfile(path, ...)
end

if os.getenv("POB2_ENGINE_TRACE_LUA") == "1" then
	io.stderr:write("pob2-lua: bytecode\n")
end
main()
return true
