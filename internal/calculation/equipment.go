package calculation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (s *inputs) readEquipment() {
	items := make(map[int]build.Item)
	for _, item := range s.model.Items {
		items[item.ID] = item
	}
	used := make(map[int]bool)
	for _, slot := range sortedKeys(s.model.Slots) {
		id := s.model.Slots[slot]
		if id == 0 {
			continue
		}
		path := "Items/" + slot
		if (s.model.WeaponSet == 1 && strings.HasSuffix(slot, " Swap")) ||
			(s.model.WeaponSet == 2 && (slot == "Weapon 1" || slot == "Weapon 2")) {
			s.issue("retained_inactive_item", path, "сохранён предмет неактивного варианта оружия")
			continue
		}
		kind := ""
		switch slot {
		case "Ring 1", "Ring 2":
			kind = "Ring"
		case "Amulet":
			kind = "Amulet"
		}
		if kind == "" {
			s.issue("unsupported_slot", path, "слот пока не поддерживается")
			continue
		}
		if used[id] {
			s.issue("unsupported_item", path, "один предмет выбран в нескольких слотах")
			continue
		}
		used[id] = true
		item, ok := items[id]
		if !ok {
			s.issue("missing_input", path, "предмет отсутствует")
			continue
		}
		s.readItem(item, kind, path)
	}
}

func (s *inputs) readItem(item build.Item, kind, path string) {
	var lines []string
	for _, line := range strings.Split(item.Raw, "\n") {
		if value := strings.TrimSpace(line); value != "" {
			lines = append(lines, value)
		}
	}
	if len(lines) < 2 {
		s.issue("unsupported_item", path, "неполный текст предмета")
		return
	}
	rarity := strings.ToUpper(lines[0])
	if rarity != "RARITY: RARE" && rarity != "RARITY: NORMAL" && rarity != "RARITY: MAGIC" {
		s.issue("unsupported_item", path, "поддерживаются обычные, волшебные и редкие предметы")
		return
	}
	baseIndex := -1
	for i := 1; i < len(lines) && i <= 2; i++ {
		if base, ok := s.engine.data.Item(lines[i]); ok {
			if base.SlotType != kind {
				s.issue("unsupported_item", path, "база несовместима со слотом")
				return
			}
			s.items = append(s.items, base)
			baseIndex = i
			break
		}
	}
	if baseIndex < 0 {
		s.issue("unsupported_item", path, "неподдерживаемая база предмета")
		return
	}
	for i, line := range lines[baseIndex+1:] {
		if strings.HasPrefix(line, "LevelReq: ") {
			value, err := strconv.Atoi(strings.TrimPrefix(line, "LevelReq: "))
			if err == nil && value > s.items[len(s.items)-1].RequiredLevel {
				s.items[len(s.items)-1].RequiredLevel = value
			}
		}
		if s.itemMetadata(line, path) {
			continue
		}
		s.addLine(line, fmt.Sprintf("%s/line/%d", path, baseIndex+i+2))
	}
}

func (s *inputs) itemMetadata(line, path string) bool {
	for _, key := range []string{"Item Level: ", "LevelReq: ", "Implicits: "} {
		if !strings.HasPrefix(line, key) {
			continue
		}
		value, err := strconv.Atoi(strings.TrimPrefix(line, key))
		if err != nil || value < 0 || value > 100 {
			s.issue("unsupported_item", path, "неверные метаданные предмета")
		}
		return true
	}
	return false
}
