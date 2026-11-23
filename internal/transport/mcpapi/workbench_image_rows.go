package mcpapi

import (
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

var imageTitles = map[string]string{
	"overview": "Сводка сборки", "tree": "Дерево пассивных умений", "stats": "Характеристики",
	"skills": "Умения и камни", "items": "Предметы", "config": "Условия расчёта",
	"notes": "Заметки", "warnings": "Предупреждения", "node": "Пассивное умение",
	"item": "Предмет", "all": "Полный отчёт о сборке",
}

type imageRow struct {
	text  string
	style string
}

type imageData struct {
	Report  map[string]any `json:"report"`
	Tooltip map[string]any `json:"tooltip"`
	Scene   imageScene     `json:"scene"`
	Tree    imageTree      `json:"tree"`
}

func (d imageData) info() map[string]any {
	return object(object(d.Report["snapshot"])["buildInfo"])
}

func (d imageData) rows(options imageOptions) ([]imageRow, error) {
	rows, err := d.section(options.view, options.id)
	if err != nil {
		return nil, err
	}
	filtered := rows[:0]
	query := strings.ToLower(options.query)
	for _, row := range rows {
		if strings.Contains(strings.ToLower(row.text), query) {
			filtered = append(filtered, row)
		}
	}
	if len(filtered) == 0 {
		filtered = append(filtered, imageRow{"По заданному фильтру ничего не найдено.", "muted"})
	}
	return filtered, nil
}

func (d imageData) section(view string, id int) ([]imageRow, error) {
	switch view {
	case "overview":
		return d.overviewRows(), nil
	case "stats":
		return mappingRows("Показатели исходного расчёта PoB", object(d.Report["stats"])), nil
	case "skills":
		return rowsForSkills(d.Report["groups"]), nil
	case "items", "item":
		return rowsForItems(d.Report, id)
	case "config":
		return configurationRows(d.Report), nil
	case "notes":
		return noteRows(d.Report["notes"]), nil
	case "warnings":
		return append([]imageRow{{"Предупреждения и неподдерживаемые модификаторы", "heading"}}, d.warnings()...), nil
	case "node":
		return d.nodeRows(id), nil
	default:
		return d.fullRows(), nil
	}
}

func (d imageData) fullRows() []imageRow {
	var rows []imageRow
	for _, section := range []string{"overview", "stats", "skills", "items", "config", "notes"} {
		part, _ := d.section(section, 0)
		rows = append(rows, part...)
	}
	rows = append(rows, mappingRows("Активные наборы", object(d.Report["activeSets"]))...)
	for _, slot := range objects(d.Report["slots"]) {
		rows = append(rows, imageRow{fmt.Sprintf("%s: предмет #%s", display(slot["name"]), display(slot["itemId"])), ""})
	}
	return append(rows, imageRow{"Назначенные узлы", "heading"}, imageRow{display(d.Report["allocatedNodes"]), ""})
}

func (d imageData) overviewRows() []imageRow {
	info := d.info()
	main := "не выбрано"
	groups := object(info["groups"])
	for _, group := range objects(groups["entries"]) {
		if group["id"] == groups["selected"] {
			main = display(group["label"])
		}
	}
	rows := []imageRow{{"Основное умение: " + main, "heading"}}
	stats := object(d.Report["stats"])
	for _, name := range []string{"TotalDPS", "Life", "EnergyShield", "Mana", "Spirit", "SpiritUnreserved", "Armour", "Evasion",
		"EvadeChance", "BlockChance", "FireResist", "ColdResist", "LightningResist", "ChaosResist", "Str", "Dex", "Int"} {
		if value, ok := stats[name]; ok {
			rows = append(rows, imageRow{name + ": " + display(value), ""})
		}
	}
	rows = append(rows, imageRow{"Очки дерева: " + display(d.Report["pointDisplay"]), "muted"})
	config := object(d.Report["config"])
	for _, name := range []string{"enemyIsBoss", "enemyLevel", "conditionEnemyChilled", "conditionUsingFlask", "resistancePenalty"} {
		if value, ok := config[name]; ok {
			rows = append(rows, imageRow{name + ": " + display(value), ""})
		}
	}
	return append(rows, d.warnings()...)
}

func (d imageData) warnings() []imageRow {
	rows := []imageRow{}
	for _, value := range array(d.Report["warnings"]) {
		rows = append(rows, imageRow{display(value), "warning"})
	}
	for _, group := range objects(d.Report["groups"]) {
		for _, gem := range objects(group["gems"]) {
			if warning := display(gem["error"]); warning != "" {
				rows = append(rows, imageRow{display(gem["name"]) + ": " + warning, "warning"})
			}
		}
	}
	for _, item := range objects(d.Report["items"]) {
		rows = append(rows, modifierRows(display(item["name"])+": ", item["unsupported"])...)
	}
	for _, block := range objects(d.Report["customMods"]) {
		rows = append(rows, modifierRows("Пользовательский модификатор: ", block["unparsed"])...)
	}
	return rows
}

func modifierRows(prefix string, raw any) []imageRow {
	rows := []imageRow{}
	for _, line := range objects(raw) {
		rows = append(rows, imageRow{prefix + display(line["line"]), "warning"})
	}
	return rows
}

func rowsForSkills(raw any) []imageRow {
	rows := []imageRow{}
	for _, group := range objects(raw) {
		rows = append(rows, imageRow{fmt.Sprintf("Группа #%s: %s", display(group["id"]), display(group["label"])), "heading"})
		for _, gem := range objects(group["gems"]) {
			rows = append(rows, imageRow{fmt.Sprintf("%s · уровень %s · качество %s%%",
				display(gem["name"]), display(gem["level"]), display(gem["quality"])), ""})
			if err := display(gem["error"]); err != "" {
				rows = append(rows, imageRow{err, "warning"})
			}
		}
	}
	return rows
}

func rowsForItems(report map[string]any, id int) ([]imageRow, error) {
	var rows []imageRow
	for _, item := range objects(report["items"]) {
		if id == 0 || number(item["id"], 0) == id {
			rows = append(rows, itemRows(report, item)...)
		}
	}
	if id != 0 && len(rows) == 0 {
		return nil, fmt.Errorf("предмет с ID %d отсутствует в сборке", id)
	}
	return rows, nil
}

func itemRows(report, item map[string]any) []imageRow {
	rows := []imageRow{{fmt.Sprintf("#%s: %s · %s",
		display(item["id"]), display(item["name"]), display(item["base"])), "heading"}}
	for _, slot := range objects(report["slots"]) {
		if slot["itemId"] == item["id"] {
			rows = append(rows, imageRow{"Слот: " + display(slot["name"]), "muted"})
		}
	}
	for line := range strings.SplitSeq(display(item["raw"]), "\n") {
		rows = append(rows, imageRow{line, ""})
	}
	return append(rows, modifierRows("Не поддерживается: ", item["unsupported"])...)
}

func configurationRows(report map[string]any) []imageRow {
	rows := mappingRows("Установленные параметры расчёта", object(report["config"]))
	rows = append(rows, mappingRows("Подсказки PoB", object(report["configPlaceholders"]))...)
	for _, block := range objects(report["customMods"]) {
		rows = append(rows, imageRow{display(block["title"]), "heading"}, imageRow{display(block["text"]), ""})
	}
	return rows
}

func mappingRows(title string, values map[string]any) []imageRow {
	rows := []imageRow{{title, "heading"}}
	for _, key := range slices.Sorted(maps.Keys(values)) {
		rows = append(rows, imageRow{key + ": " + display(values[key]), ""})
	}
	return rows
}

func noteRows(raw any) []imageRow {
	return []imageRow{{"Заметки автора сборки", "heading"}, {display(raw), ""}}
}

func (d imageData) nodeRows(id int) []imageRow {
	rows := []imageRow{{fmt.Sprintf("Узел #%d", id), "heading"}}
	for _, line := range objects(d.Tooltip["lines"]) {
		rows = append(rows, imageRow{display(line["text"]), ""})
	}
	return append(rows, imageRow{"Маршрут: " + display(d.Tooltip["path"]), "muted"}, imageRow{display(d.Tooltip["note"]), ""})
}

var imageMarkup = regexp.MustCompile(`\^x[0-9a-fA-F]{6}|\^[0-9]|[\x00-\x08\x0b\x0c\x0e-\x1f]`)

func display(value any) string {
	switch value := value.(type) {
	case nil:
		return ""
	case string:
		return imageMarkup.ReplaceAllString(value, "")
	case bool:
		if value {
			return "да"
		}
		return "нет"
	default:
		data, _ := json.Marshal(value)
		return string(data)
	}
}

func object(value any) map[string]any {
	result, _ := value.(map[string]any)
	return result
}

func number(value any, fallback int) int {
	if n, ok := value.(float64); ok {
		return int(n)
	}
	return fallback
}

func array(value any) []any {
	result, _ := value.([]any)
	return result
}

func objects(value any) []map[string]any {
	list := array(value)
	result := make([]map[string]any, 0, len(list))
	for _, item := range list {
		result = append(result, object(item))
	}
	return result
}
