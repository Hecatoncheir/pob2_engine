//go:build !windows

package pobwindow

import (
	"path/filepath"
	"runtime"
)

func platformCandidates() []string {
	if runtime.GOOS != "linux" {
		return nil
	}
	prefix, err := WinePrefix("", "")
	if err != nil {
		return nil
	}
	root := filepath.Join(prefix, "drive_c")
	patterns := []string{
		"Program Files/Path of Building*",
		"Program Files (x86)/Path of Building*",
		"users/*/AppData/Roaming/Path of Building*",
		"users/*/AppData/Local/Path of Building*",
	}
	var paths []string
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(root, filepath.FromSlash(pattern)))
		paths = append(paths, matches...)
	}
	return paths
}
