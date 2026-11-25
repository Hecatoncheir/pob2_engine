package pob

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func TestTreeGeometryReplyUnpacksScene(t *testing.T) {
	const reply = `{"version":"0_5","nodes":[{"id":17}],"scene":{
		"width":1800,"scale":0.25,
		"command_strings":["TreeData/atlas.png","ALPHA","TreeData/other.png"],
		"command_values":[
			3,2,7,1,-1.25,2,3,4,5,6,7,8,0,0,1,0,1,1,0,1,1,0.5,0.25,1,
			1,2,3,0,1,2,3,4,5,6,7,8,0,0,1,0,1,1,0,1,1,1,1,1]}}`
	var data treeGeometryReply
	if err := json.Unmarshal([]byte(reply), &data); err != nil {
		t.Fatal(err)
	}
	geometry, err := data.unpack()
	if err != nil {
		t.Fatal(err)
	}
	want := []build.TreeDrawCommand{
		{
			File: "TreeData/other.png", Blend: "ALPHA", Index: 7, Clamp: true,
			Positions: []float64{-1.25, 2, 3, 4, 5, 6, 7, 8},
			UV:        []float64{0, 0, 1, 0, 1, 1, 0, 1},
			Color:     []float64{1, 0.5, 0.25, 1},
		},
		{
			File: "TreeData/atlas.png", Blend: "ALPHA", Index: 3,
			Positions: []float64{1, 2, 3, 4, 5, 6, 7, 8},
			UV:        []float64{0, 0, 1, 0, 1, 1, 0, 1},
			Color:     []float64{1, 1, 1, 1},
		},
	}
	if geometry.Version != "0_5" || len(geometry.Nodes) != 1 || geometry.Nodes[0].ID != 17 ||
		geometry.Scene == nil || geometry.Scene.Width != 1800 || geometry.Scene.Scale != 0.25 {
		t.Fatalf("неверная геометрия: %#v", geometry)
	}
	if !reflect.DeepEqual(geometry.Scene.Commands, want) {
		t.Fatalf("команды = %#v, ожидались %#v", geometry.Scene.Commands, want)
	}
}

func TestTreeGeometryReplyAcceptsSceneWithoutPackedCommands(t *testing.T) {
	var data treeGeometryReply
	if err := json.Unmarshal([]byte(`{"scene":{"commands":[{"file":"asset.png"}]}}`), &data); err != nil {
		t.Fatal(err)
	}
	geometry, err := data.unpack()
	if err != nil {
		t.Fatal(err)
	}
	want := []build.TreeDrawCommand{{File: "asset.png"}}
	if geometry.Scene == nil || !reflect.DeepEqual(geometry.Scene.Commands, want) {
		t.Fatalf("неверная сцена: %#v", geometry.Scene)
	}
}

func TestTreeSceneReplyRejectsInvalidPackedCommands(t *testing.T) {
	row := make([]float64, treeDrawRowSize)
	row[0], row[1] = 1, 2
	for _, test := range []struct {
		name   string
		values []float64
	}{
		{"неполная строка", row[:len(row)-1]},
		{"неизвестная строка", append([]float64{3}, row[1:]...)},
		{"недопустимый флаг", append([]float64{1, 2, 0, 2}, row[4:]...)},
	} {
		t.Run(test.name, func(t *testing.T) {
			scene := treeSceneReply{Strings: []string{"asset.png", "ALPHA"}, Values: test.values}
			if err := scene.unpackCommands(); err == nil {
				t.Fatal("повреждённый ответ принят")
			}
		})
	}
}
