//go:build !linux || !cgo

package pobwindow

import "fmt"

func ControlWineWindow(action, executable, ownership string) (bool, error) {
	return false, fmt.Errorf("управление окном Wine доступно в сборке Linux с X11")
}
