package pobwindow

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type buildDirectory struct {
	Directory string `json:"directory"`
	Wine      *struct {
		Prefix           string `json:"prefix"`
		WorkingDirectory string `json:"workingDirectory"`
	} `json:"wine"`
}

func (c *Client) BuildDirectory(ctx context.Context) (string, error) {
	value, err := c.Call(ctx, "getBuildDirectory", nil)
	if err != nil {
		return "", err
	}
	var result buildDirectory
	if err := json.Unmarshal(value, &result); err != nil {
		return "", err
	}
	if result.Directory == "" {
		return "", directoryUnavailable()
	}
	if result.Wine == nil {
		return result.Directory, nil
	}
	prefix, err := WinePrefix(c.directory, result.Wine.Prefix)
	if err != nil {
		return "", err
	}
	return mountedDirectory(result.Directory, result.Wine.WorkingDirectory, prefix)
}

func WinePrefix(installation, reported string) (string, error) {
	if filepath.IsAbs(reported) {
		return filepath.Clean(reported), nil
	}
	for current := installation; current != ""; current = parentDirectory(current) {
		if filepath.Base(current) == "drive_c" {
			return filepath.Dir(current), nil
		}
	}
	if prefix := os.Getenv("WINEPREFIX"); prefix != "" {
		return filepath.Abs(prefix)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".wine"), nil
}

func parentDirectory(current string) string {
	parent := filepath.Dir(current)
	if parent == current || parent == "." {
		return ""
	}
	return parent
}

func mountedDirectory(location, working, prefix string) (string, error) {
	if filepath.IsAbs(location) {
		return filepath.Clean(location), nil
	}
	drive, relative := absoluteWinePath(location, working)
	root := mappedDrive(prefix, drive)
	if root == "" {
		return "", directoryUnavailable()
	}
	for name := range strings.SplitSeq(relative, "/") {
		root = existingName(root, name)
	}
	return root, nil
}

func absoluteWinePath(location, working string) (string, string) {
	location = strings.ReplaceAll(location, "\\", "/")
	if !wineAbsolute(location) && working != "" && !strings.Contains(location, ":") {
		location = strings.TrimSuffix(strings.ReplaceAll(working, "\\", "/"), "/") + "/" + location
	}
	if !wineAbsolute(location) {
		return "", ""
	}
	return strings.ToLower(location[:2]), strings.TrimPrefix(path.Clean(location[2:]), "/")
}

func wineAbsolute(path string) bool {
	if len(path) < 3 || path[1:3] != ":/" {
		return false
	}
	letter := path[0] | 32
	return 'a' <= letter && letter <= 'z'
}

func mappedDrive(prefix, drive string) string {
	if drive == "" {
		return ""
	}
	devices := filepath.Join(prefix, "dosdevices")
	if info, err := os.Stat(devices); err == nil && info.IsDir() {
		mapping, err := filepath.EvalSymlinks(filepath.Join(devices, drive))
		if err == nil {
			return mapping
		}
		return ""
	}
	switch drive {
	case "c:":
		return filepath.Join(prefix, "drive_c")
	case "z:":
		return "/"
	default:
		return ""
	}
}

func pathForWine(location, prefix string) (string, error) {
	location, err := filepath.EvalSymlinks(location)
	if err != nil {
		return "", err
	}
	if universal := mappedDrive(prefix, "z:"); universal == "/" {
		return "Z:" + filepath.ToSlash(location), nil
	}
	drive, relative := matchingDrive(location, prefix)
	if drive == "" {
		return "", &BridgeError{"WINE_PATH_UNAVAILABLE", "Каталог моста недоступен через диски префикса Wine"}
	}
	return strings.ToUpper(drive) + "/" + filepath.ToSlash(relative), nil
}

func matchingDrive(location, prefix string) (string, string) {
	var best, relative string
	length := 0
	for letter := byte('a'); letter <= 'z'; letter++ {
		drive := string(letter) + ":"
		root := mappedDrive(prefix, drive)
		if len(root) <= length {
			continue
		}
		inside, err := filepath.Rel(root, location)
		if err == nil && inside != ".." && !strings.HasPrefix(inside, ".."+string(filepath.Separator)) {
			best, relative, length = drive, inside, len(root)
		}
	}
	return best, relative
}

func existingName(parent, name string) string {
	exact := filepath.Join(parent, name)
	if _, err := os.Stat(exact); err == nil {
		return exact
	}
	entries, _ := os.ReadDir(parent)
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), name) {
			return filepath.Join(parent, entry.Name())
		}
	}
	return exact
}

func directoryUnavailable() error {
	return &BridgeError{"BUILD_DIRECTORY_UNAVAILABLE", "Не удалось определить папку сборок PoB"}
}
