package pob

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var treePathCallback = sync.OnceValue(func() uintptr {
	return windows.NewCallback(func(input *uint32, inputSize uint32, output *uint32, outputSize uint32) uintptr {
		return uintptr(solvePathWords(input, inputSize, output, outputSize))
	})
})

func runLua(library, program string) error {
	return runLuaProgram(library, treePathProgram(program, treePathCallback))
}

func runLuaProgram(library, program string) error {
	handle, err := windows.LoadLibraryEx(library, 0, windows.LOAD_LIBRARY_SEARCH_DLL_LOAD_DIR|windows.LOAD_LIBRARY_SEARCH_DEFAULT_DIRS)
	if err != nil {
		return err
	}
	dll := &windows.DLL{Handle: handle, Name: library}
	defer dll.Release()
	names := []string{"luaL_newstate", "luaL_openlibs", "luaL_loadstring", "lua_pcall", "lua_tolstring", "lua_close"}
	procs, err := luaProcedures(dll, names)
	if err != nil {
		return err
	}
	return executeLua(procs, program)
}

func luaProcedures(dll *windows.DLL, names []string) ([]*windows.Proc, error) {
	procs := make([]*windows.Proc, 0, len(names))
	for _, name := range names {
		p, err := dll.FindProc(name)
		if err != nil {
			return nil, err
		}
		procs = append(procs, p)
	}
	return procs, nil
}

func executeLua(p []*windows.Proc, program string) error {
	code, err := windows.BytePtrFromString(program)
	if err != nil {
		return err
	}
	state, _, _ := p[0].Call()
	if state == 0 {
		return fmt.Errorf("не удалось создать состояние Lua")
	}
	defer p[5].Call(state)
	p[1].Call(state)
	status, _, _ := p[2].Call(state, uintptr(unsafe.Pointer(code)))
	runtime.KeepAlive(code)
	if int32(status) == 0 {
		status, _, _ = p[3].Call(state, 0, 0, 0)
	}
	if int32(status) == 0 {
		return nil
	}
	return luaError(p[4], state, status)
}

func luaError(getString *windows.Proc, state, status uintptr) error {
	index := -1
	var size uintptr
	address, _, _ := getString.Call(state, uintptr(index), uintptr(unsafe.Pointer(&size)))
	if address == 0 || size == 0 {
		return fmt.Errorf("Lua: код %d", status)
	}
	message := make([]byte, min(size, 4096))
	if err := windows.ReadProcessMemory(windows.CurrentProcess(), address, &message[0], uintptr(len(message)), nil); err != nil {
		return fmt.Errorf("Lua: код %d: %w", status, err)
	}
	return fmt.Errorf("Lua: %s", message)
}
