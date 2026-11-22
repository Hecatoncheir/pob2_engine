//go:build linux || darwin

package discovery

import "os"

func openRecord(path string) (*os.File, error) { return os.Open(path) }
