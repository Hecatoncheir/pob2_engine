package maintenance

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
	"golang.org/x/sys/windows"
)

func replaceExecutable(source, target string) error {
	err := privatefs.Replace(source, target)
	if !canMoveRunningImage(err) {
		return err
	}
	// Windows может удерживать запущенный образ до выхода процесса.
	dir, err := os.MkdirTemp(filepath.Dir(target), ".pob2-engine-previous-*")
	if err != nil {
		return err
	}
	defer os.Remove(dir)
	backup := filepath.Join(dir, filepath.Base(target))
	if err := os.Rename(target, backup); err != nil {
		return err
	}
	if err := privatefs.Replace(source, target); err != nil {
		if restore := os.Rename(backup, target); restore != nil {
			return fmt.Errorf("замена: %w. Восстановление: %w. Предыдущий бинарник: %s", err, restore, backup)
		}
		return err
	}
	// Если старый процесс ещё активен, файл сохраняется для ручного удаления после выхода.
	_ = os.Remove(backup)
	return nil
}

func canMoveRunningImage(err error) bool {
	return errors.Is(err, windows.ERROR_ACCESS_DENIED) || errors.Is(err, windows.ERROR_SHARING_VIOLATION) ||
		errors.Is(err, windows.ERROR_USER_MAPPED_FILE)
}
