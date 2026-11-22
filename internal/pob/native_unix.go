//go:build (linux || darwin) && cgo

package pob

/*
#cgo linux LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>

static int run_lua(const char *library, const char *program, char *message, size_t size) {
	void *dl = dlopen(library, RTLD_NOW | RTLD_GLOBAL);
	if (!dl) {
		snprintf(message, size, "%s", dlerror());
		return 1;
	}
	void *(*create)(void) = dlsym(dl, "luaL_newstate");
	void (*openlibs)(void *) = dlsym(dl, "luaL_openlibs");
	int (*load)(void *, const char *) = dlsym(dl, "luaL_loadstring");
	int (*call)(void *, int, int, int) = dlsym(dl, "lua_pcall");
	const char *(*string)(void *, int, size_t *) = dlsym(dl, "lua_tolstring");
	void (*close)(void *) = dlsym(dl, "lua_close");
	if (!create || !openlibs || !load || !call || !string || !close) {
		snprintf(message, size, "%s", "Неполный Lua C API");
		dlclose(dl);
		return 1;
	}
	void *state = create();
	if (!state) {
		snprintf(message, size, "%s", "Не удалось создать состояние Lua");
		dlclose(dl);
		return 1;
	}
	openlibs(state);
	int status = load(state, program);
	if (status == 0) status = call(state, 0, 0, 0);
	if (status != 0) {
		const char *error = string(state, -1, NULL);
		snprintf(message, size, "%s", error ? error : "Неизвестная ошибка Lua");
	}
	close(state);
	dlclose(dl);
	return status;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func runLua(library, program string) error {
	path, code := C.CString(library), C.CString(program)
	defer C.free(unsafe.Pointer(path))
	defer C.free(unsafe.Pointer(code))
	var message [4096]C.char
	if C.run_lua(path, code, &message[0], C.size_t(len(message))) != 0 {
		return fmt.Errorf("Lua: %s", C.GoString(&message[0]))
	}
	return nil
}
