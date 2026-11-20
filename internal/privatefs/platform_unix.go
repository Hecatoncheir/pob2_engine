//go:build linux || darwin

package privatefs

import "os"

func protectDir(path string) error { return os.Chmod(path, 0700) }
