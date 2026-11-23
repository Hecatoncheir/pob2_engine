package pobwindow

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Installations struct {
	Paths     []string `json:"paths"`
	Preferred string   `json:"preferred,omitempty"`
}

func Discover(state string) Installations {
	config, _, _ := loadBridgeConfig(state)
	explicit := []string{os.Getenv("POB_DIR"), config.InstallDir}
	paths := installationPaths(append(explicit, platformCandidates()...))
	result := Installations{Paths: paths}
	if preferred := installationPaths(explicit); len(preferred) != 0 {
		result.Preferred = preferred[0]
	}
	return result
}

func installationPaths(candidates []string) []string {
	paths := []string{}
	seen := map[string]bool{}
	for _, candidate := range candidates {
		root := resolveInstallation(candidate)
		if root == "" || seen[installationKey(root)] {
			continue
		}
		paths = append(paths, root)
		seen[installationKey(root)] = true
	}
	return paths
}

func resolveInstallation(candidate string) string {
	if candidate == "" {
		return ""
	}
	directory, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return ""
	}
	if isFile(directory) {
		directory = filepath.Dir(directory)
	}
	root, err := filepath.Abs(directory)
	if err != nil {
		return ""
	}
	if _, err := launchFile(root); err != nil {
		return ""
	}
	return root
}

func installationKey(root string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(root)
	}
	return root
}
