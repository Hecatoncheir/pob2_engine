package gamedata

import "testing"

func TestRuntimeCatalogDoesNotChooseAnAmbiguousGem(t *testing.T) {
	value := RuntimeCatalog{TreeVersion: "0_5", Classes: []Class{{ID: 7}}, Items: []Item{{Name: "Base"}},
		Nodes: []Node{{ID: 1}}, Gems: []Gem{{ID: "first", Name: "Shared"}, {ID: "second", Name: "Shared"}}}
	data, err := NewRuntimeCatalog(value, "local", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, found := data.Gem("Shared"); found {
		t.Fatal("неоднозначное имя выбрало случайный камень")
	}
	if gem, found := data.Gem("second"); !found || gem.ID != "second" {
		t.Fatal("точный идентификатор не разрешился")
	}
}

func TestRuntimeCatalogKeepsIndependentSnapshots(t *testing.T) {
	value := RuntimeCatalog{TreeVersion: "0_5", Classes: []Class{{ID: 7}},
		Items: []Item{{Name: "Base", Slots: []string{"Body Armour"}}}, Nodes: []Node{{ID: 1, Neighbors: []int{2}}},
		Gems: []Gem{{ID: "gem", Name: "Name", Levels: []GemLevel{{Level: 1}}, Qualities: []int{0, 20}}}}
	data, err := NewRuntimeCatalog(value, "local", "", "")
	if err != nil {
		t.Fatal(err)
	}
	value.Gems[0].Qualities[0], value.Items[0].Slots[0], value.Nodes[0].Neighbors[0] = 99, "Changed", 3
	first, _ := data.Gem("gem")
	first.Levels[0].Level, first.Qualities[0] = 99, 99
	second, _ := data.Gem("gem")
	if second.Levels[0].Level != 1 || second.Qualities[0] != 0 || data.Items()[0].Slots[0] != "Body Armour" || data.Nodes()[0].Neighbors[0] != 2 {
		t.Fatal("изменение полученной копии испортило общий каталог")
	}
}
