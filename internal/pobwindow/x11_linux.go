//go:build linux && cgo

package pobwindow

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
#include <string.h>

typedef void Display;
typedef unsigned long Window;
typedef unsigned long Atom;

typedef struct {
	int type;
	unsigned long serial;
	int send_event;
	Display *display;
	Window window;
	Atom message_type;
	int format;
	long data[5];
} Message;

typedef union {
	Message message;
	long padding[24];
} Event;

typedef struct {
	void *library;
	Display *display;
	Display *(*open_display)(const char *);
	int (*close_display)(Display *);
	Window (*root)(Display *);
	int (*screen)(Display *);
	Atom (*atom)(Display *, const char *, int);
	int (*property)(Display *, Window, Atom, long, long, int, Atom, Atom *, int *, unsigned long *, unsigned long *, unsigned char **);
	int (*free)(void *);
	int (*sync)(Display *, int);
	int (*iconify)(Display *, Window, int);
	int (*raise)(Display *, Window);
	int (*send)(Display *, Window, int, long, Event *);
	void *(*set_error_handler)(int (*)(Display *, void *));
} X11;

static int ignore_x11_error(Display *display, void *event) { return 0; }

static void close_x11(X11 *x11) {
	if (x11->display) x11->close_display(x11->display);
	if (x11->library) dlclose(x11->library);
	free(x11);
}

static X11 *open_x11(void) {
	X11 *x11 = calloc(1, sizeof(*x11));
	if (!x11) return NULL;
	x11->library = dlopen("libX11.so.6", RTLD_NOW | RTLD_LOCAL);
	if (!x11->library) goto failed;
#define LOAD(field, name) do { x11->field = dlsym(x11->library, name); if (!x11->field) goto failed; } while (0)
	LOAD(open_display, "XOpenDisplay");
	LOAD(close_display, "XCloseDisplay");
	LOAD(root, "XDefaultRootWindow");
	LOAD(screen, "XDefaultScreen");
	LOAD(atom, "XInternAtom");
	LOAD(property, "XGetWindowProperty");
	LOAD(free, "XFree");
	LOAD(sync, "XSync");
	LOAD(iconify, "XIconifyWindow");
	LOAD(raise, "XMapRaised");
	LOAD(send, "XSendEvent");
	LOAD(set_error_handler, "XSetErrorHandler");
#undef LOAD
	x11->display = x11->open_display(NULL);
	if (!x11->display) goto failed;
	x11->set_error_handler(ignore_x11_error);
	return x11;
failed:
	close_x11(x11);
	return NULL;
}

static unsigned long values(X11 *x11, Window window, const char *name, Atom type, unsigned long limit, unsigned long *out) {
	Atom actual;
	int format;
	unsigned long count, remaining;
	unsigned char *data = NULL;
	int result = x11->property(x11->display, window, x11->atom(x11->display, name, 0), 0, limit, 0, type,
		&actual, &format, &count, &remaining, &data);
	unsigned long copied = 0;
	if (result == 0 && data && actual == type && format == 32 && count <= limit) {
		memcpy(out, data, count * sizeof(unsigned long));
		copied = count;
	}
	if (data) x11->free(data);
	return copied;
}

static int windows(X11 *x11, Window *out, unsigned long limit) {
	Window root = x11->root(x11->display), manager;
	if (!values(x11, root, "_NET_SUPPORTING_WM_CHECK", 33, 1, &manager)) return -1;
	return values(x11, root, "_NET_CLIENT_LIST", 33, limit, out);
}

static unsigned long owner_pid(X11 *x11, Window window) {
	unsigned long pid = 0;
	values(x11, window, "_NET_WM_PID", 6, 1, &pid);
	return pid;
}

static int request_close(X11 *x11, Window window) {
	Atom delete = x11->atom(x11->display, "WM_DELETE_WINDOW", 0);
	Atom protocols[64];
	unsigned long size = values(x11, window, "WM_PROTOCOLS", 4, 64, protocols);
	for (unsigned long index = 0; index < size; index++) {
		if (protocols[index] != delete) continue;
		Event event = {0};
		event.message.type = 33;
		event.message.send_event = 1;
		event.message.display = x11->display;
		event.message.window = window;
		event.message.message_type = x11->atom(x11->display, "WM_PROTOCOLS", 0);
		event.message.format = 32;
		event.message.data[0] = delete;
		return x11->send(x11->display, window, 0, 0, &event);
	}
	return 0;
}

static int window_action(X11 *x11, Window window, int action) {
	switch (action) {
	case 1: return x11->iconify(x11->display, window, x11->screen(x11->display));
	case 2: return x11->raise(x11->display, window);
	case 3: return request_close(x11, window);
	default: return 1;
	}
}

static void flush_x11(X11 *x11) { x11->sync(x11->display, 0); }
*/
import "C"

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const wineOwnerVariable = "POB2_UI_WINE_OWNER"

var wineActions = map[string]C.int{"check": 0, "minimize": 1, "restore": 2, "close": 3}

func ControlWineWindow(action, executable, ownership string) (bool, error) {
	code, ok := wineActions[action]
	if !ok || executable == "" || code != 0 && len(ownership) < 32 {
		return false, fmt.Errorf("неверные параметры управления окном Wine")
	}
	display := C.open_x11()
	if display == nil {
		return false, fmt.Errorf("не удалось открыть дисплей X11")
	}
	defer C.close_x11(display)
	var windows [4096]C.Window
	count := C.windows(display, &windows[0], C.ulong(len(windows)))
	if count < 0 {
		return false, fmt.Errorf("не найден оконный менеджер X11 или XWayland")
	}
	matched := controlOwnedWindows(display, windows[:int(count)], executable, ownership, code)
	C.flush_x11(display)
	return matched, nil
}

func controlOwnedWindows(display *C.X11, windows []C.Window, executable, ownership string, code C.int) bool {
	matched := false
	for _, window := range windows {
		pid := uint64(C.owner_pid(display, window))
		if !ownedWineWindow(pid, executable, ownership, code == 0) {
			continue
		}
		if C.window_action(display, window, code) != 0 {
			matched = true
		}
	}
	return matched
}

func ownedWineWindow(pid uint64, executable, token string, probe bool) bool {
	if pid == 0 {
		return false
	}
	process := filepath.Join("/proc", strconv.FormatUint(pid, 10))
	line, err := os.ReadFile(filepath.Join(process, "cmdline"))
	if err != nil {
		return false
	}
	command, _, _ := bytes.Cut(line, []byte{0})
	name := path.Base(strings.ReplaceAll(string(command), "\\", "/"))
	if !strings.EqualFold(name, executable) {
		return false
	}
	return probe || processOwnsWindow(process, token)
}

func processOwnsWindow(process, token string) bool {
	environment, err := os.ReadFile(filepath.Join(process, "environ"))
	if err != nil {
		return false
	}
	for entry := range bytes.SplitSeq(environment, []byte{0}) {
		if string(entry) == wineOwnerVariable+"="+token {
			return true
		}
	}
	return false
}
