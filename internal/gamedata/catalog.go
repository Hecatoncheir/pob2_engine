// Package gamedata загружает проверяемый набор данных без выполнения Lua.
package gamedata

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"regexp"
	"slices"
)

//go:embed files/*.json
var embedded embed.FS

type Class struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Strength     int    `json:"strength"`
	Dexterity    int    `json:"dexterity"`
	Intelligence int    `json:"intelligence"`
	StartNode    int    `json:"start_node"`
}

type GemLevel struct {
	Level                int     `json:"level"`
	RequiredLevel        int     `json:"required_level"`
	RequiredIntelligence int     `json:"required_intelligence"`
	FireMin              float64 `json:"fire_min"`
	FireMax              float64 `json:"fire_max"`
	CritChance           float64 `json:"crit_chance"`
	Mana                 float64 `json:"mana"`
}

type Stat struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Gem struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	SkillID       string     `json:"skill_id"`
	Support       bool       `json:"support"`
	CastTime      float64    `json:"cast_time"`
	Family        string     `json:"family"`
	Levels        []GemLevel `json:"levels"`
	ConstantStats []Stat     `json:"constant_stats"`
}

type Item struct {
	Name          string `json:"name"`
	SlotType      string `json:"slot_type"`
	Implicit      string `json:"implicit"`
	RequiredLevel int    `json:"required_level"`
}

type Node struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Stats     []string `json:"stats"`
	Neighbors []int    `json:"neighbors"`
}

type Constants struct {
	LifePerLevel  float64 `json:"life_per_level"`
	ManaPerLevel  float64 `json:"mana_per_level"`
	MaxResistance float64 `json:"base_maximum_all_resistances_%"`
	CritBonus     float64 `json:"base_critical_hit_damage_bonus"`
	ManaRegen     float64 `json:"mana_regeneration_base"`
}

type catalog struct {
	Classes   []Class   `json:"classes"`
	Gems      []Gem     `json:"gems"`
	Items     []Item    `json:"items"`
	Nodes     []Node    `json:"nodes"`
	Constants Constants `json:"constants"`
}

type Manifest struct {
	Schema      int               `json:"schema"`
	ID          string            `json:"id"`
	GameVersion *string           `json:"game_version"`
	TreeVersion string            `json:"tree_version"`
	PoBCommit   string            `json:"pob_commit"`
	Files       map[string]string `json:"files"`
	Exporters   map[string]string `json:"exporters"`
	Origin      struct {
		Repository string            `json:"repository"`
		Commit     string            `json:"commit"`
		Files      map[string]string `json:"files"`
	} `json:"origin"`
	Rights map[string]string `json:"rights"`
}

// Dataset после загрузки не изменяется; методы возвращают независимые срезы.
type Dataset struct {
	id, tree, commit string
	catalog          catalog
}

func Embedded() (*Dataset, error) {
	files, err := fs.Sub(embedded, "files")
	if err != nil {
		return nil, err
	}
	return Load(files)
}

func decode(data []byte, target any) error {
	if len(data) > 2<<20 {
		return fmt.Errorf("файл набора превышает 2 МиБ")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("лишние данные JSON")
	}
	return nil
}

var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func Load(files fs.FS) (*Dataset, error) {
	data, err := fs.ReadFile(files, "manifest.json")
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	if err := decode(data, &manifest); err != nil {
		return nil, err
	}
	if err := validateManifest(manifest); err != nil {
		return nil, err
	}
	data, err = fs.ReadFile(files, "catalog.json")
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != manifest.Files["catalog.json"] {
		return nil, fmt.Errorf("нарушена SHA-256 каталога")
	}
	dataset := &Dataset{id: manifest.ID, tree: manifest.TreeVersion, commit: manifest.PoBCommit}
	if err := decode(data, &dataset.catalog); err != nil {
		return nil, err
	}
	if err := dataset.validate(); err != nil {
		return nil, err
	}
	return dataset, nil
}

func validateManifest(m Manifest) error {
	if m.Schema != 1 || m.ID == "" || m.TreeVersion != "0_5" || !commitPattern.MatchString(m.PoBCommit) {
		return fmt.Errorf("несовместимый манифест набора")
	}
	if m.Origin.Commit != m.PoBCommit || m.Origin.Repository == "" || len(m.Origin.Files) == 0 || len(m.Rights) == 0 {
		return fmt.Errorf("неполное происхождение набора")
	}
	if len(m.Files) != 1 || !hashPattern.MatchString(m.Files["catalog.json"]) || len(m.Exporters) != 2 {
		return fmt.Errorf("неполный состав набора")
	}
	for _, name := range []string{"tool/export_data.lua", "tool/export_data.py"} {
		if !hashPattern.MatchString(m.Exporters[name]) {
			return fmt.Errorf("неверная SHA-256 экспортёра")
		}
	}
	for _, digest := range m.Origin.Files {
		if !hashPattern.MatchString(digest) {
			return fmt.Errorf("неверная SHA-256 источника")
		}
	}
	return nil
}

func (d *Dataset) ID() string           { return d.id }
func (d *Dataset) TreeVersion() string  { return d.tree }
func (d *Dataset) PoBCommit() string    { return d.commit }
func (d *Dataset) Constants() Constants { return d.catalog.Constants }
func (d *Dataset) Classes() []Class     { return slices.Clone(d.catalog.Classes) }
func (d *Dataset) Class(id int) (Class, bool) {
	for _, c := range d.catalog.Classes {
		if c.ID == id {
			return c, true
		}
	}
	return Class{}, false
}
func (d *Dataset) Gem(id string) (Gem, bool) {
	for _, g := range d.catalog.Gems {
		if g.ID == id || g.SkillID == id || g.Name == id {
			g.Levels = slices.Clone(g.Levels)
			g.ConstantStats = slices.Clone(g.ConstantStats)
			return g, true
		}
	}
	return Gem{}, false
}
func (d *Dataset) Item(name string) (Item, bool) {
	for _, i := range d.catalog.Items {
		if i.Name == name {
			return i, true
		}
	}
	return Item{}, false
}
func (d *Dataset) Node(id int) (Node, bool) {
	for _, n := range d.catalog.Nodes {
		if n.ID == id {
			n.Stats = slices.Clone(n.Stats)
			n.Neighbors = slices.Clone(n.Neighbors)
			return n, true
		}
	}
	return Node{}, false
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
