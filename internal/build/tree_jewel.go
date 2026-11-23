package build

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// JewelEdit меняет самоцвет по ID гнезда активной спецификации.
type JewelEdit struct {
	NodeID int    `json:"node_id"`
	ItemID int    `json:"item_id,omitempty"`
	Raw    string `json:"raw,omitempty"`
}

func (edit JewelEdit) validate() error {
	if edit.NodeID < 1 || edit.NodeID > 1000000 {
		return fmt.Errorf("неверный идентификатор гнезда самоцвета")
	}
	if edit.ItemID < 0 || edit.ItemID > 1000000 || len(edit.Raw) > 65536 || !utf8.ValidString(edit.Raw) {
		return fmt.Errorf("неверный идентификатор или текст самоцвета")
	}
	if edit.ItemID > 0 && edit.Raw != "" {
		return fmt.Errorf("укажите либо текст самоцвета, либо его ID")
	}
	if edit.Raw != "" && strings.TrimSpace(edit.Raw) == "" {
		return fmt.Errorf("пустой текст самоцвета")
	}
	return nil
}

func (d *Document) editJewel(edit JewelEdit, runtime bool) error {
	if !runtime {
		return fmt.Errorf("правка самоцвета доступна с оригинальным Lua PoB")
	}
	id, err := d.jewelItemID(edit)
	if err != nil {
		return err
	}
	return d.setTreeJewel(edit.NodeID, id)
}

func (d *Document) jewelItemID(edit JewelEdit) (string, error) {
	raw := strings.TrimSpace(edit.Raw)
	if edit.ItemID == 0 && raw == "" {
		return "0", nil
	}
	items, _, err := d.equipmentSet(true)
	if err != nil {
		return "", err
	}
	if edit.ItemID == 0 {
		return newEquipmentItem(items, raw)
	}
	id := strconv.Itoa(edit.ItemID)
	for _, item := range items.children("Item") {
		if existing, _ := strconv.Atoi(item.attr("id", "")); existing == edit.ItemID {
			return id, nil
		}
	}
	return "", fmt.Errorf("предмет %d отсутствует в сборке", edit.ItemID)
}

func readTreeJewels(spec *element) map[string]int {
	sockets, err := spec.child("Sockets")
	if err != nil || sockets == nil {
		return nil
	}
	items := make(map[string]int)
	for _, socket := range sockets.children("Socket") {
		nodeID, nodeErr := integer(socket.attr("nodeId", ""), "Socket/nodeId", 1, 1000000)
		itemID, itemErr := integer(socket.attr("itemId", ""), "Socket/itemId", 1, 1000000)
		if nodeErr == nil && itemErr == nil {
			items[strconv.Itoa(nodeID)] = itemID
		}
	}
	return items
}

func (d *Document) setTreeJewel(nodeID int, itemID string) error {
	tree, err := d.root.child("Tree")
	if err != nil {
		return err
	}
	if tree == nil {
		return fmt.Errorf("в сборке нет дерева пассивов")
	}
	spec, _, err := activeTreeSpec(tree)
	if err != nil {
		return err
	}
	sockets, err := spec.child("Sockets")
	if err != nil || (sockets == nil && itemID == "0") {
		return err
	}
	if sockets == nil {
		sockets = newElement("Sockets")
		spec.parts = append(spec.parts, part{element: sockets})
	}
	return updateTreeJewel(sockets, nodeID, itemID)
}

func updateTreeJewel(sockets *element, nodeID int, itemID string) error {
	var selected *element
	for _, socket := range sockets.children("Socket") {
		if id, _ := strconv.Atoi(socket.attr("nodeId", "")); id != nodeID {
			continue
		}
		if selected != nil {
			return fmt.Errorf("гнездо %d повторяется в XML", nodeID)
		}
		selected = socket
	}
	if itemID == "0" {
		if selected != nil {
			sockets.remove(selected)
		}
		return nil
	}
	if selected == nil {
		selected = newElement("Socket")
		selected.set("nodeId", strconv.Itoa(nodeID))
		sockets.parts = append(sockets.parts, part{element: selected})
	}
	selected.set("itemId", itemID)
	return nil
}
