package pob

import (
	"fmt"
	"math"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

// Вместо повторения полей и имён текстур Lua может вернуть словарь и строки по 24 числа.
const treeDrawRowSize = 24

type treeGeometryReply struct {
	build.TreeGeometry
	Scene *treeSceneReply `json:"scene"`
}

type treeSceneReply struct {
	build.TreeScene
	Strings []string  `json:"command_strings"`
	Values  []float64 `json:"command_values"`
}

func (reply *treeGeometryReply) unpack() (build.TreeGeometry, error) {
	geometry := reply.TreeGeometry
	if reply.Scene == nil {
		return geometry, nil
	}
	if err := reply.Scene.unpackCommands(); err != nil {
		return build.TreeGeometry{}, err
	}
	geometry.Scene = &reply.Scene.TreeScene
	return geometry, nil
}

func (scene *treeSceneReply) unpackCommands() error {
	if scene.Values == nil {
		return nil
	}
	if scene.Commands != nil || len(scene.Values)%treeDrawRowSize != 0 {
		return fmt.Errorf("неверное число полей команд сцены PoB")
	}
	scene.Commands = make([]build.TreeDrawCommand, 0, len(scene.Values)/treeDrawRowSize)
	for start := 0; start < len(scene.Values); start += treeDrawRowSize {
		command, err := treeDrawCommand(scene.Values[start:start+treeDrawRowSize], scene.Strings)
		if err != nil {
			return err
		}
		scene.Commands = append(scene.Commands, command)
	}
	return nil
}

func treeDrawCommand(row []float64, strings []string) (build.TreeDrawCommand, error) {
	file, fileOK := sceneString(strings, row[0])
	blend, blendOK := sceneString(strings, row[1])
	if !fileOK || !blendOK {
		return build.TreeDrawCommand{}, fmt.Errorf("неверный индекс строки в команде сцены PoB")
	}
	index, clamp := row[2], row[3]
	if index < 0 || index > 1<<32-1 || index != math.Trunc(index) || (clamp != 0 && clamp != 1) {
		return build.TreeDrawCommand{}, fmt.Errorf("неверные параметры команды сцены PoB")
	}
	return build.TreeDrawCommand{
		File: file, Blend: blend, Index: uint32(index), Clamp: clamp == 1,
		Positions: row[4:12:12], UV: row[12:20:20], Color: row[20:24:24],
	}, nil
}

func sceneString(strings []string, index float64) (string, bool) {
	if index < 1 || index > float64(len(strings)) || index != math.Trunc(index) {
		return "", false
	}
	return strings[int(index)-1], true
}
