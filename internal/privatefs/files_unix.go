//go:build linux || darwin

package privatefs

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// Replace атомарно переименовывает файл в пределах одной файловой системы.
func Replace(source, target string) error { return os.Rename(source, target) }

func lockFile(file *os.File) error {
	err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EWOULDBLOCK) {
		return ErrLocked
	}
	return err
}
