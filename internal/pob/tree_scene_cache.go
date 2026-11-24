package pob

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
)

const (
	treeSceneCacheFormat = "1"
	treeSceneCacheLimit  = 24 << 20
)

// Храним только последнюю сцену и только для проверяемой установки PoB.
type treeSceneCache struct {
	path   string
	source string
}

// EnableTreeSceneCache вызывают до первого запроса на сервере с профилем.
func (c *Calculator) EnableTreeSceneCache(directory string) {
	c.treeScene = localTreeSceneCache(c.installation, directory)
}

func localTreeSceneCache(installation Installation, directory string) treeSceneCache {
	manager, err := Local()
	if err != nil {
		return treeSceneCache{}
	}
	managed, err := manager.readManagedInstallation()
	if err != nil {
		return treeSceneCache{}
	}
	for _, candidate := range []*Installation{&managed.Installation, managed.Previous} {
		if candidate != nil && matchesTreeSceneInstallation(installation, *candidate) {
			return treeSceneCache{
				path:   filepath.Join(directory, "last-tree-scene.gz"),
				source: candidate.Version + "/" + candidate.Commit + "/" + candidate.SHA256,
			}
		}
	}
	return treeSceneCache{}
}

func matchesTreeSceneInstallation(actual, managed Installation) bool {
	return actual.Directory == managed.Directory &&
		(actual.NativeDirectory == "" || actual.NativeDirectory == managed.NativeDirectory)
}

func (c treeSceneCache) key(xml []byte) [sha256.Size]byte {
	hash := sha256.New()
	io.WriteString(hash, treeSceneCacheFormat+"/"+buildinfo.Version+"/"+c.source+"\n")
	hash.Write(xml)
	return [sha256.Size]byte(hash.Sum(nil))
}

func (c treeSceneCache) read(key [sha256.Size]byte) (build.TreeGeometry, bool) {
	file, err := os.Open(c.path)
	if err != nil {
		return build.TreeGeometry{}, false
	}
	defer file.Close()

	reader, err := gzip.NewReader(io.LimitReader(file, treeSceneCacheLimit))
	if err != nil {
		return build.TreeGeometry{}, false
	}
	defer reader.Close()
	if !bytes.Equal(reader.Extra, key[:]) {
		return build.TreeGeometry{}, false
	}
	data, err := io.ReadAll(io.LimitReader(reader, treeSceneCacheLimit+1))
	if err != nil || len(data) > treeSceneCacheLimit {
		return build.TreeGeometry{}, false
	}
	var tree build.TreeGeometry
	if err := json.Unmarshal(data, &tree); err != nil || !completeTreeScene(tree) {
		return build.TreeGeometry{}, false
	}
	return tree, true
}

func (c treeSceneCache) write(key [sha256.Size]byte, tree build.TreeGeometry) {
	if !completeTreeScene(tree) {
		return
	}
	data, err := json.Marshal(tree)
	if err != nil || len(data) > treeSceneCacheLimit {
		return
	}
	file, err := os.CreateTemp(filepath.Dir(c.path), ".tree-scene-")
	if err != nil {
		return
	}
	defer os.Remove(file.Name())
	err = writeTreeScene(file, key, data)
	if closeErr := file.Close(); err == nil && closeErr == nil {
		_ = privatefs.Replace(file.Name(), c.path)
	}
}

func writeTreeScene(file io.Writer, key [sha256.Size]byte, data []byte) error {
	writer, err := gzip.NewWriterLevel(file, gzip.BestSpeed)
	if err != nil {
		return err
	}
	writer.Extra = key[:]
	if _, err := writer.Write(data); err != nil {
		return err
	}
	return writer.Close()
}

func completeTreeScene(tree build.TreeGeometry) bool {
	return tree.Version != "" && len(tree.Nodes) != 0 &&
		tree.Scene != nil && len(tree.Scene.Commands) != 0
}
