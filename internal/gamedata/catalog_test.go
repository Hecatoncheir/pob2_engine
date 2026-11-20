package gamedata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"testing/fstest"
)

func copyFiles(t *testing.T) fstest.MapFS {
	t.Helper()
	result := fstest.MapFS{}
	for _, name := range []string{"catalog.json", "manifest.json"} {
		data, err := embedded.ReadFile("files/" + name)
		if err != nil {
			t.Fatal(err)
		}
		result[name] = &fstest.MapFile{Data: data}
	}
	return result
}

func updateCatalog(t *testing.T, files fstest.MapFS, mutate func(*catalog)) {
	t.Helper()
	var c catalog
	if err := json.Unmarshal(files["catalog.json"].Data, &c); err != nil {
		t.Fatal(err)
	}
	mutate(&c)
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	files["catalog.json"].Data = data
	var manifest Manifest
	if err := json.Unmarshal(files["manifest.json"].Data, &manifest); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	manifest.Files["catalog.json"] = hex.EncodeToString(digest[:])
	files["manifest.json"].Data, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEmbeddedAndIndependentValues(t *testing.T) {
	d, err := Embedded()
	if err != nil {
		t.Fatal(err)
	}
	gem, ok := d.Gem("FireballPlayer")
	if !ok || gem.Levels[0].DamageMin != 8 || gem.Levels[19].DamageMax != 336 {
		t.Fatalf("неверные уровни: %+v", gem)
	}
	gem.Levels[0].DamageMin = -1
	node, _ := d.Node(4739)
	node.Stats[0] = "bad"
	again, _ := d.Gem("FireballPlayer")
	nodeAgain, _ := d.Node(4739)
	if again.Levels[0].DamageMin != 8 || nodeAgain.Stats[0] == "bad" {
		t.Fatal("можно изменить загруженный набор через результат")
	}
}

func TestThreeSpellProfiles(t *testing.T) {
	d, err := Embedded()
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct{ id, element, mode string }{
		{"FireballPlayer", "Fire", "Projectile"}, {"FrostboltPlayer", "Cold", "Projectile"}, {"SparkPlayer", "Lightning", "Base"},
	} {
		gem, ok := d.Gem(sample.id)
		if !ok || gem.DamageType != sample.element || gem.Mode != sample.mode || len(gem.Levels) != 20 || gem.Levels[0].DamageMin <= 0 {
			t.Fatalf("неполный профиль %s: %+v", sample.id, gem)
		}
	}
}

func TestRejectDamagedAndIncompatibleData(t *testing.T) {
	cases := map[string]func(fstest.MapFS){
		"checksum": func(f fstest.MapFS) { f["catalog.json"].Data = append(f["catalog.json"].Data, '\n') },
		"schema": func(f fstest.MapFS) {
			var manifest Manifest
			if err := json.Unmarshal(f["manifest.json"].Data, &manifest); err != nil {
				t.Fatal(err)
			}
			manifest.Schema = 3
			data, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			f["manifest.json"].Data = data
		},
		"missing":   func(f fstest.MapFS) { delete(f, "catalog.json") },
		"duplicate": func(f fstest.MapFS) { updateCatalog(t, f, func(c *catalog) { c.Gems = append(c.Gems, c.Gems[0]) }) },
		"edge": func(f fstest.MapFS) {
			updateCatalog(t, f, func(c *catalog) { c.Nodes[0].Neighbors = append(c.Nodes[0].Neighbors, 999999) })
		},
		"class start": func(f fstest.MapFS) { updateCatalog(t, f, func(c *catalog) { c.Classes[0].StartNode = 999999 }) },
		"level":       func(f fstest.MapFS) { updateCatalog(t, f, func(c *catalog) { c.Gems[1].Levels[0].DamageMin = -5 }) },
		"damage type": func(f fstest.MapFS) { updateCatalog(t, f, func(c *catalog) { c.Gems[0].DamageType = "Unknown" }) },
		"mode":        func(f fstest.MapFS) { updateCatalog(t, f, func(c *catalog) { c.Gems[0].Mode = "Explosion" }) },
		"damage range": func(f fstest.MapFS) {
			updateCatalog(t, f, func(c *catalog) { c.Gems[0].Levels[0].DamageMax = c.Gems[0].Levels[0].DamageMin - 1 })
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			files := copyFiles(t)
			mutate(files)
			if _, err := Load(files); err == nil {
				t.Fatal("повреждённый набор принят")
			}
		})
	}
}
