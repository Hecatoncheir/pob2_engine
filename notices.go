// Package notices предоставляет оригинальные лицензионные уведомления в бинарнике.
package notices

import (
	"embed"
	"io/fs"
)

//go:embed LICENSE THIRD_PARTY_NOTICES.md all:.licenses internal/pob/bridge/LICENSE
var files embed.FS

// Files возвращает тексты уведомлений по исходным путям в репозитории.
func Files() (map[string]string, error) {
	result := make(map[string]string)
	err := fs.WalkDir(files, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := files.ReadFile(path)
		if err != nil {
			return err
		}
		result[path] = string(data)
		return nil
	})
	return result, err
}
