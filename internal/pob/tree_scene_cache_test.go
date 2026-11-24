package pob

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func TestTreeSceneCacheReusesOnlyMatchingScene(t *testing.T) {
	cache := treeSceneCache{path: filepath.Join(t.TempDir(), "scene.gz"), source: "pob-a"}
	xml := []byte("<PathOfBuilding/>")
	key := cache.key(xml)
	tree := build.TreeGeometry{
		Version: "0_5",
		Nodes:   []build.TreeGeometryNode{{ID: 17, Allocated: true}},
		Scene:   &build.TreeScene{Commands: []build.TreeDrawCommand{{File: "asset.png"}}},
	}
	cache.write(key, tree)
	if found, ok := cache.read(key); !ok || !reflect.DeepEqual(found, tree) {
		t.Fatalf("сцена из кэша = %#v, найдена = %v", found, ok)
	}
	if _, ok := cache.read(cache.key([]byte("<PathOfBuilding><Tree/></PathOfBuilding>"))); ok {
		t.Fatal("другая сборка получила прежнюю сцену")
	}
	newPoB := cache
	newPoB.source = "pob-b"
	if _, ok := newPoB.read(newPoB.key(xml)); ok {
		t.Fatal("новая версия PoB получила прежнюю сцену")
	}
	file, err := os.Stat(cache.path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(cache.path, file.Size()-2); err != nil {
		t.Fatal(err)
	}
	if _, ok := cache.read(key); ok {
		t.Fatal("повреждённая сцена принята")
	}
	cache.write(key, tree)
	if _, ok := cache.read(key); !ok {
		t.Fatal("повреждённая запись не восстановлена")
	}
}
