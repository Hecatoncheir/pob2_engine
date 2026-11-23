package application

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
)

type CatalogQuery struct {
	Kind   string `json:"kind" jsonschema:"gems, classes, items или passives"`
	Query  string `json:"query,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// CatalogEntry описывает покрытие движка, а не полный игровой объект.
type CatalogEntry struct {
	Kind            string   `json:"kind"`
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Aliases         []string `json:"aliases"`
	Levels          []int    `json:"levels"`
	Qualities       []int    `json:"qualities"`
	Modes           []string `json:"modes"`
	Support         bool     `json:"support"`
	Family          string   `json:"family"`
	Slots           []string `json:"slots"`
	ClassIDs        []int    `json:"class_ids"`
	Coverage        []string `json:"coverage"`
	AttributeChoice bool     `json:"attribute_choice,omitempty"`
}

type CatalogPage struct {
	DataID      string         `json:"data_id"`
	DataSHA256  string         `json:"data_sha256"`
	PoBCommit   string         `json:"pob_commit"`
	TreeVersion string         `json:"tree_version"`
	Kind        string         `json:"kind"`
	Query       string         `json:"query"`
	Offset      int            `json:"offset"`
	Limit       int            `json:"limit"`
	Total       int            `json:"total"`
	Entries     []CatalogEntry `json:"entries"`
}

type CatalogCommands interface {
	Catalog(context.Context, CatalogQuery) (CatalogPage, error)
	PreviewSupport(context.Context, PreviewSupport) (SupportPreview, error)
}

func (s *Service) Catalog(ctx context.Context, query CatalogQuery) (CatalogPage, error) {
	if err := ctx.Err(); err != nil {
		return CatalogPage{}, err
	}
	query, err := validateCatalogQuery(query)
	if err != nil {
		return CatalogPage{}, err
	}
	data, err := s.editData(ctx)
	if err != nil {
		return CatalogPage{}, err
	}
	entries := catalogEntries(data, query.Kind)
	entries = slices.DeleteFunc(entries, func(entry CatalogEntry) bool { return !matchesCatalog(entry, query.Query) })
	slices.SortFunc(entries, func(a, b CatalogEntry) int { return strings.Compare(a.ID, b.ID) })
	total := len(entries)
	start := min(query.Offset, total)
	end := min(start+query.Limit, total)
	return CatalogPage{DataID: data.ID(), DataSHA256: data.SHA256(), PoBCommit: data.PoBCommit(), TreeVersion: data.TreeVersion(),
		Kind: query.Kind, Query: query.Query, Offset: query.Offset, Limit: query.Limit, Total: total, Entries: entries[start:end]}, ctx.Err()
}

func validateCatalogQuery(query CatalogQuery) (CatalogQuery, error) {
	if !slices.Contains([]string{"gems", "classes", "items", "passives"}, query.Kind) {
		return query, failure(InvalidArgument, "kind: нужны gems, classes, items или passives")
	}
	if !utf8.ValidString(query.Query) || len(query.Query) > 256 || query.Offset < 0 || query.Offset > 10000 || query.Limit < 0 || query.Limit > 100 {
		return query, failure(InvalidArgument, "поиск: UTF-8 до 256 байт, offset 0–10000, limit 1–100 (0 означает 20)")
	}
	query.Query = strings.TrimSpace(query.Query)
	if query.Limit == 0 {
		query.Limit = 20
	}
	return query, nil
}

func matchesCatalog(entry CatalogEntry, query string) bool {
	query = strings.ToLower(query)
	for _, value := range append([]string{entry.ID, entry.Name}, entry.Aliases...) {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	return false
}

func emptyEntry(kind, id, name string) CatalogEntry {
	return CatalogEntry{Kind: kind, ID: id, Name: name, Aliases: []string{}, Levels: []int{}, Qualities: []int{}, Modes: []string{}, Slots: []string{}, ClassIDs: []int{}, Coverage: []string{}}
}

func gemEntry(gem gamedata.Gem) CatalogEntry {
	entry := emptyEntry("gems", gem.ID, gem.Name)
	entry.Aliases = []string{gem.GameID, gem.VariantID, gem.SkillID, gem.Name}
	entry.Support, entry.Family, entry.Qualities = gem.Support, gem.Family, []int{0}
	for _, level := range gem.Levels {
		entry.Levels = append(entry.Levels, level.Level)
	}
	if len(gem.Qualities) > 0 {
		entry.Qualities = gem.Qualities
		entry.Coverage = []string{"Камень из активной установки PoB; расчёт и предупреждения возвращает оригинальный Lua."}
		if gem.Mode != "" {
			entry.Modes = []string{gem.Mode}
		}
		return entry
	}
	entry.Coverage = []string{"Только statSetIndex=1, без осквернения; одна включённая группа, одно активное умение."}
	if gem.Support {
		entry.Coverage = append(entry.Coverage, "Поддержка Fireball/Frostbolt/Spark; уровень 1, качество 0; без повтора семейства.")
		return entry
	}
	entry.Modes = []string{gem.Mode}
	entry.Coverage = append(entry.Coverage, "Только попадания в одну цель; без взрывов, состояний, DoT, Infusion и повторных попаданий Spark.")
	return entry
}

func catalogEntries(data *gamedata.Dataset, kind string) []CatalogEntry {
	switch kind {
	case "gems":
		return gemEntries(data)
	case "classes":
		return classEntries(data)
	case "items":
		return itemEntries(data)
	case "passives":
		return passiveEntries(data)
	}
	return []CatalogEntry{}
}

func gemEntries(data *gamedata.Dataset) []CatalogEntry {
	entries := []CatalogEntry{}
	for _, gem := range data.Gems() {
		entries = append(entries, gemEntry(gem))
	}
	return entries
}

func classEntries(data *gamedata.Dataset) []CatalogEntry {
	entries := []CatalogEntry{}
	for _, class := range data.Classes() {
		entry := emptyEntry("classes", strconv.Itoa(class.ID), class.Name)
		entry.ClassIDs = []int{class.ID}
		entry.Coverage = []string{"Базовые характеристики, без возвышений; старт дерева " + strconv.Itoa(class.StartNode) + ". Пассивы выбирайте отдельно из каталога."}
		if data.IsRuntime() {
			entry.Coverage = []string{"Базовый класс оригинального PoB; редактор возвышений пока отсутствует."}
		}
		entries = append(entries, entry)
	}
	return entries
}

func itemEntries(data *gamedata.Dataset) []CatalogEntry {
	entries := []CatalogEntry{}
	for _, item := range data.Items() {
		entry := itemEntry(item)
		if data.IsRuntime() {
			entry = runtimeItemEntry(item)
		}
		entries = append(entries, entry)
	}
	return entries
}

func passiveEntries(data *gamedata.Dataset) []CatalogEntry {
	entries := []CatalogEntry{}
	for _, node := range data.Nodes() {
		entries = append(entries, passiveEntry(data, node))
	}
	return entries
}

func itemEntry(item gamedata.Item) CatalogEntry {
	entry := emptyEntry("items", item.Name, item.Name)
	entry.Slots = []string{"Amulet"}
	if item.SlotType == "Ring" {
		entry.Slots = []string{"Ring 1", "Ring 2"}
	}
	entry.Coverage = []string{"Только эта база и ограниченный разбор свойств; уникальные механики, сокеты и прочие свойства не подразумеваются поддержанными.", "Необходимый уровень базы: " + strconv.Itoa(item.RequiredLevel) + "; собственное свойство: " + item.Implicit}
	return entry
}

func passiveEntry(data *gamedata.Dataset, node gamedata.Node) CatalogEntry {
	entry := emptyEntry("passives", strconv.Itoa(node.ID), node.Name)
	entry.AttributeChoice = node.AttributeChoice
	for _, class := range data.Classes() {
		if class.StartNode == node.ID || !data.IsRuntime() && class.ID == 7 && len(node.Neighbors) > 0 {
			entry.ClassIDs = append(entry.ClassIDs, class.ID)
		}
	}
	entry.Coverage = []string{"Только перечисленные классы; нужны связь со стартом и доступные очки без наград заданий. Это не полное дерево."}
	if data.IsRuntime() {
		entry.Coverage = []string{"Узел активного дерева PoB; применимость выделенных узлов проверяет оригинальный PoB."}
	}
	return entry
}

func runtimeItemEntry(item gamedata.Item) CatalogEntry {
	entry := emptyEntry("items", item.Name, item.Name)
	if len(item.Slots) > 0 {
		entry.Slots = item.Slots
	}
	entry.Coverage = []string{"База из активной установки PoB; тип: " + item.SlotType + ". Для правки используйте слот выбранного набора XML и полный текст предмета."}
	return entry
}
