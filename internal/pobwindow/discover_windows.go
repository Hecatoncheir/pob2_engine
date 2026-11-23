package pobwindow

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func platformCandidates() []string {
	var paths []string
	for _, key := range []string{"APPDATA", "LOCALAPPDATA", "ProgramFiles", "ProgramFiles(x86)"} {
		if root := os.Getenv(key); root != "" {
			paths = append(paths, filepath.Join(root, "Path of Building Community (PoE2)"), filepath.Join(root, "Path of Building (PoE2)"))
		}
	}
	for _, hive := range []registry.Key{registry.CURRENT_USER, registry.LOCAL_MACHINE} {
		paths = append(paths, registryInstallations(hive, registry.WOW64_64KEY)...)
		paths = append(paths, registryInstallations(hive, registry.WOW64_32KEY)...)
	}
	return paths
}

func registryInstallations(hive registry.Key, view uint32) []string {
	root, err := registry.OpenKey(hive, `Software\Microsoft\Windows\CurrentVersion\Uninstall`, registry.READ|view)
	if err != nil {
		return nil
	}
	defer root.Close()
	entries, err := root.ReadSubKeyNames(-1)
	if err != nil {
		return nil
	}
	var paths []string
	for _, name := range entries {
		if location := registryLocation(root, name); location != "" {
			paths = append(paths, location)
		}
	}
	return paths
}

func registryLocation(root registry.Key, name string) string {
	entry, err := registry.OpenKey(root, name, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer entry.Close()
	title, _, err := entry.GetStringValue("DisplayName")
	if err != nil || !strings.Contains(title, "Path of Building") || !strings.Contains(title, "PoE2") {
		return ""
	}
	location, _, _ := entry.GetStringValue("InstallLocation")
	return strings.Trim(location, `"`)
}
