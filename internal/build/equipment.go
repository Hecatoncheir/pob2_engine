package build

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (edit EquipmentEdit) validate() error {
	if edit.Slot == "" || len(edit.Slot) > 128 || !utf8.ValidString(edit.Slot) || len(edit.Raw) > 65536 || !utf8.ValidString(edit.Raw) {
		return fmt.Errorf("неверный слот или текст предмета")
	}
	if edit.ItemID < 0 || edit.ItemID > 1000000 || (edit.ItemID > 0 && edit.Raw != "") {
		return fmt.Errorf("укажите либо текст предмета, либо его ID")
	}
	return nil
}

func (d *Document) editEquipment(edit EquipmentEdit, runtime bool) error {
	items, set, err := d.equipmentSet(runtime)
	if err != nil {
		return err
	}
	slot, err := equipmentSlot(set, edit.Slot, runtime)
	if err != nil {
		return err
	}
	id, err := equipmentItemID(items, edit.ItemID, edit.Raw)
	if err != nil {
		return err
	}
	slot.set("itemId", id)
	return nil
}

func equipmentItemID(items *element, itemID int, raw string) (string, error) {
	if itemID == 0 {
		return newEquipmentItem(items, strings.TrimSpace(raw))
	}
	value := strconv.Itoa(itemID)
	for _, item := range items.children("Item") {
		if existing, _ := strconv.Atoi(item.attr("id", "")); existing == itemID {
			return value, nil
		}
	}
	return "", fmt.Errorf("предмет %d отсутствует в сборке", itemID)
}

func (d *Document) equipmentSet(runtime bool) (*element, *element, error) {
	items, err := d.root.child("Items")
	if err != nil {
		return nil, nil, err
	}
	if items == nil {
		if !runtime {
			return nil, nil, fmt.Errorf("нет секции Items")
		}
		items, set := newElement("Items"), newElement("ItemSet")
		items.set("activeItemSet", "1")
		set.set("id", "1")
		items.parts = append(items.parts, part{element: set})
		d.root.parts = append(d.root.parts, part{element: items})
		return items, set, nil
	}
	set, _, err := selectedByID(items, "ItemSet", "activeItemSet")
	return items, set, err
}

func equipmentSlot(set *element, name string, runtime bool) (*element, error) {
	for _, slot := range set.children("Slot") {
		if slot.attr("name", "") == name {
			return slot, nil
		}
	}
	if !runtime {
		return nil, fmt.Errorf("слот %s отсутствует в выбранном наборе", name)
	}
	// Новый слот в копии должен подтвердить Lua до её использования.
	slot := newElement("Slot")
	slot.set("name", name)
	set.parts = append(set.parts, part{element: slot})
	return slot, nil
}

func newEquipmentItem(items *element, raw string) (string, error) {
	if raw == "" {
		return "0", nil
	}
	previous := items.children("Item")
	if len(previous) >= 512 {
		return "", fmt.Errorf("слишком много предметов")
	}
	maximum := 0
	for _, node := range previous {
		id, err := integer(node.attr("id", ""), "Item/id", 1, 1000000)
		if err != nil {
			return "", err
		}
		maximum = max(maximum, id)
	}
	if maximum >= 1000000 {
		return "", fmt.Errorf("слишком много предметов")
	}
	id := strconv.Itoa(maximum + 1)
	item := newElement("Item")
	item.set("id", id)
	item.parts = []part{{text: raw}}
	items.parts = append(items.parts, part{element: item})
	return id, nil
}
