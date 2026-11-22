package privatefs

import (
	"errors"
	"os"
)

var ErrLocked = errors.New("файл занят другим процессом")

// TryLock удерживает исключительную блокировку до закрытия файла.
// Файл блокировки сохраняется, чтобы другие процессы всегда открывали тот же файл.
func TryLock(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := lockFile(file); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}
