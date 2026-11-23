package mcpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"sync"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
)

type workbenchAssets struct {
	reader *pob.Calculator
	scene  map[string]build.TreeTextureQuery
	remote *gemImages
}

func newWorkbenchAssets(installation pob.Installation) (*workbenchAssets, error) {
	reader, err := pob.NewCalculator(installation)
	if err != nil {
		return nil, err
	}
	remote, err := newGemImages()
	if err != nil {
		return nil, err
	}
	return &workbenchAssets{reader: reader, scene: map[string]build.TreeTextureQuery{}, remote: remote}, nil
}

func (a *workbenchAssets) prepare(ctx context.Context, name string, value pob.WorkbenchResult) (pob.WorkbenchResult, error) {
	switch name {
	case "pob_get_tree", "pob_tree_hover", "pob_node_power":
		return a.prepareScene(value)
	case "pob_tree_icons":
		return a.prepareIcons(ctx, value, a.treeIcon)
	case "pob_gem_icons":
		return a.prepareIcons(ctx, value, a.gemIcon)
	default:
		return value, nil
	}
}

func (a *workbenchAssets) prepareScene(value pob.WorkbenchResult) (pob.WorkbenchResult, error) {
	var response map[string]json.RawMessage
	if err := json.Unmarshal(value.Result, &response); err != nil {
		return value, err
	}
	if len(response["scene"]) == 0 {
		return value, nil
	}
	scene, err := a.sceneCommands(response["scene"])
	if err != nil {
		return value, err
	}
	response["scene"] = scene
	return workbenchResult(value.Revision, response)
}

func (a *workbenchAssets) sceneCommands(raw json.RawMessage) (json.RawMessage, error) {
	var scene map[string]json.RawMessage
	if err := json.Unmarshal(raw, &scene); err != nil {
		return nil, err
	}
	var commands []drawCommand
	if err := json.Unmarshal(scene["commands"], &commands); err != nil {
		return nil, err
	}
	wire := make([]sceneCommand, 0, len(commands))
	for _, command := range commands {
		asset, err := a.sceneCommand(command)
		if err != nil {
			return nil, err
		}
		wire = append(wire, asset)
	}
	data, err := json.Marshal(wire)
	if err != nil {
		return nil, err
	}
	scene["commands"] = data
	return json.Marshal(scene)
}

type drawCommand struct {
	build.TreeDrawCommand
	Mask json.RawMessage `json:"mask"`
}

type sceneCommand struct {
	Texture   string    `json:"texture"`
	Positions []float64 `json:"positions"`
	UV        []float64 `json:"uv"`
	Color     []float64 `json:"color"`
	Blend     string    `json:"blend"`
	Clamp     bool      `json:"clamp"`
}

func (a *workbenchAssets) sceneCommand(command drawCommand) (sceneCommand, error) {
	if len(command.Mask) > 0 && string(command.Mask) != "null" {
		return sceneCommand{}, fmt.Errorf("маска готовой сцены PoB не поддерживается")
	}
	query := build.TreeTextureQuery{File: command.File, Index: max(1, command.Index), TargetSize: 512}
	if err := query.Validate(); err != nil {
		return sceneCommand{}, err
	}
	token := fmt.Sprintf("%x", sha256.Sum256(fmt.Appendf(nil, "%s\x00%d", query.File, query.Index)))
	a.scene[token] = query
	return sceneCommand{token, roundCoordinates(command.Positions, 1e6), roundCoordinates(command.UV, 1e8), command.Color, command.Blend, command.Clamp}, nil
}

func roundCoordinates(values []float64, scale float64) []float64 {
	for index, value := range values {
		values[index] = math.Round(value*scale) / scale
	}
	return values
}

func (a *workbenchAssets) texture(ctx context.Context, token string) (pob.WorkbenchResult, error) {
	query, ok := a.scene[token]
	if !ok {
		return pob.WorkbenchResult{}, fmt.Errorf("текстура не принадлежит готовой сцене дерева")
	}
	texture, err := a.reader.ReadTreeTexture(ctx, query)
	if err != nil {
		return pob.WorkbenchResult{}, err
	}
	return workbenchResult(0, map[string]any{"token": token, "png": texture.PNG})
}

type iconAsset struct {
	ID    string    `json:"id"`
	Key   string    `json:"key"`
	File  string    `json:"file"`
	Index uint32    `json:"index"`
	UV    []float64 `json:"uv"`
	Error string    `json:"error"`
}

func (a *workbenchAssets) prepareIcons(ctx context.Context, value pob.WorkbenchResult, render func(context.Context, iconAsset) map[string]any) (pob.WorkbenchResult, error) {
	var result struct {
		Icons []iconAsset `json:"icons"`
	}
	if err := json.Unmarshal(value.Result, &result); err != nil {
		return value, err
	}
	icons := renderWorkbenchIcons(ctx, result.Icons, render)
	return workbenchResult(value.Revision, map[string]any{"icons": icons})
}

func renderWorkbenchIcons(ctx context.Context, assets []iconAsset, render func(context.Context, iconAsset) map[string]any) []map[string]any {
	icons := make([]map[string]any, len(assets))
	var workers sync.WaitGroup
	for offset := range min(4, len(assets)) {
		workers.Go(func() {
			for index := offset; index < len(assets); index += 4 {
				icons[index] = render(ctx, assets[index])
			}
		})
	}
	workers.Wait()
	return icons
}

func (a *workbenchAssets) treeIcon(ctx context.Context, asset iconAsset) map[string]any {
	result := map[string]any{"key": asset.Key}
	if asset.Error != "" {
		result["error"] = asset.Error
		return result
	}
	image, err := a.reader.ReadTreeTexture(ctx, build.TreeTextureQuery{File: asset.File, Index: asset.Index, TargetSize: 64})
	if err == nil {
		result["png"], err = cropTexture(image, asset.UV)
	}
	if err != nil {
		result["error"] = err.Error()
	}
	return result
}

func cropTexture(texture build.TreeTexture, uv []float64) ([]byte, error) {
	original, err := png.Decode(bytes.NewReader(texture.PNG))
	if err != nil {
		return nil, err
	}
	bounds, err := iconBounds(uv, original.Bounds())
	if err != nil {
		return nil, err
	}
	destination := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(destination, destination.Bounds(), original, bounds.Min, draw.Src)
	var encoded bytes.Buffer
	err = png.Encode(&encoded, destination)
	return encoded.Bytes(), err
}

func iconBounds(uv []float64, original image.Rectangle) (image.Rectangle, error) {
	if len(uv) != 4 {
		return image.Rectangle{}, fmt.Errorf("неверные координаты иконки PoB")
	}
	bounds := image.Rect(
		int(math.Round(uv[0]*float64(original.Dx()))), int(math.Round(uv[1]*float64(original.Dy()))),
		int(math.Round(uv[2]*float64(original.Dx()))), int(math.Round(uv[3]*float64(original.Dy()))),
	)
	if bounds.Empty() || !bounds.In(original) || uv[0] >= uv[2] || uv[1] >= uv[3] {
		return image.Rectangle{}, fmt.Errorf("область иконки находится вне атласа PoB")
	}
	return bounds, nil
}

func (a *workbenchAssets) gemIcon(ctx context.Context, asset iconAsset) map[string]any {
	result := map[string]any{"id": asset.ID}
	data, mime, source, err := a.gemImage(ctx, asset)
	if err != nil {
		result["error"] = err.Error()
		return result
	}
	result["data"], result["mimeType"], result["source"] = data, mime, source
	return result
}

func (a *workbenchAssets) gemImage(ctx context.Context, asset iconAsset) ([]byte, string, string, error) {
	if asset.Error != "" {
		data, origin, err := a.remote.read(ctx, asset.ID)
		return data, "image/webp", origin, err
	}
	image, err := a.reader.ReadGemTexture(ctx, asset.File, asset.Index)
	if len(image.PNG) > maxGemImage {
		return nil, "", "", fmt.Errorf("иконка камня превышает допустимый размер")
	}
	return image.PNG, "image/png", "original-pob", err
}
