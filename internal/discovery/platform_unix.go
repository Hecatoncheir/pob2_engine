//go:build linux || darwin

package discovery

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func openRecord(path string) (*os.File, error)  { return os.Open(path) }
func replaceRecord(source, target string) error { return os.Rename(source, target) }

func lockFile(file *os.File) error {
	err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EWOULDBLOCK) {
		return ErrLocked
	}
	return err
}
