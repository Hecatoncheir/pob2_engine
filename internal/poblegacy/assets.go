package poblegacy

import (
	"embed"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed all:assets
var assets embed.FS

func Read(name string) ([]byte, error) {
	return assets.ReadFile(path.Join("assets", name))
}

func Unpack(directory, group string) error {
	root := path.Join("assets", group)
	return fs.WalkDir(assets, root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative := strings.TrimPrefix(name, "assets")
		target := filepath.Join(directory, filepath.FromSlash(strings.TrimPrefix(relative, "/")))
		if entry.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		data, err := assets.ReadFile(name)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0600)
	})
}
