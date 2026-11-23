package mcpapi

import (
	"cmp"
	"context"
	"fmt"
	"image"
	"image/color"
	"maps"
	"math"
	"slices"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

type imageScene struct {
	Width    int           `json:"width"`
	Height   int           `json:"height"`
	Scale    float64       `json:"scale"`
	OffsetX  float64       `json:"offsetX"`
	OffsetY  float64       `json:"offsetY"`
	Commands []drawCommand `json:"commands"`
}

type imageTree struct {
	Nodes []imageTreeNode `json:"nodes"`
}

type imageTreeNode struct {
	X         float64
	Y         float64
	Name      string
	Allocated bool
}

func (a *workbenchAssets) paintTree(ctx context.Context, canvas *workbenchCanvas, options imageOptions, data imageData) (map[string]any, error) {
	scene := data.Scene
	frame := image.Rect(28, 114, 28+scene.Width, 114+scene.Height).Intersect(canvas.pixels.Bounds())
	if frame.Empty() || len(scene.Commands) == 0 {
		return nil, fmt.Errorf("исходный PoB вернул пустую сцену дерева")
	}
	pixels := canvas.pixels.SubImage(frame).(*image.RGBA)
	if err := a.applyTreeCommands(ctx, pixels, scene.Commands); err != nil {
		return nil, err
	}
	allocated := 0
	for _, node := range data.Tree.Nodes {
		if node.Allocated {
			allocated++
			if options.labels {
				labelTreeNode(canvas, frame, scene, node)
			}
		}
	}
	return map[string]any{"nodeCount": len(data.Tree.Nodes), "allocatedNodeCount": allocated,
		"commands": len(scene.Commands), "source": "original-lua"}, nil
}

func labelTreeNode(canvas *workbenchCanvas, frame image.Rectangle, scene imageScene, node imageTreeNode) {
	point := frame.Min.Add(image.Pt(int(node.X*scene.Scale+scene.OffsetX), int(node.Y*scene.Scale+scene.OffsetY)))
	if point.In(frame) {
		canvas.text(point.X+6, point.Y-4, node.Name, 13, imageText)
	}
}

func (a *workbenchAssets) applyTreeCommands(ctx context.Context, frame *image.RGBA, commands []drawCommand) error {
	textures, err := a.loadSceneTextures(ctx, commands)
	if err != nil {
		return err
	}
	for _, command := range commands {
		if err := ctx.Err(); err != nil {
			return err
		}
		paintTreeCommand(frame, textures[commandTexture(command)], command.TreeDrawCommand)
	}
	return nil
}

func (a *workbenchAssets) loadSceneTextures(ctx context.Context, commands []drawCommand) (map[build.TreeTextureQuery]*image.NRGBA, error) {
	keys, err := sceneTextures(commands)
	if err != nil {
		return nil, err
	}
	textures := make(map[build.TreeTextureQuery]*image.NRGBA, len(keys))
	pixels := 0
	for _, query := range keys {
		texture, err := a.reader.ReadTreeImage(ctx, query)
		if err != nil {
			return nil, err
		}
		pixels += texture.Bounds().Dx() * texture.Bounds().Dy()
		if pixels > 64<<20 {
			return nil, fmt.Errorf("текстуры дерева превышают допустимый объём изображения")
		}
		textures[query] = texture
	}
	return textures, nil
}

func sceneTextures(commands []drawCommand) ([]build.TreeTextureQuery, error) {
	unique := make(map[build.TreeTextureQuery]struct{})
	for _, command := range commands {
		if len(command.Mask) != 0 && string(command.Mask) != "null" {
			return nil, fmt.Errorf("маска готовой сцены PoB не поддерживается")
		}
		unique[commandTexture(command)] = struct{}{}
	}
	// Все слои одного DDS читаются подряд. Порядок отрисовки остаётся исходным.
	return slices.SortedFunc(maps.Keys(unique), func(a, b build.TreeTextureQuery) int {
		return cmp.Compare(a.File, b.File)
	}), nil
}

func commandTexture(command drawCommand) build.TreeTextureQuery {
	return build.TreeTextureQuery{File: command.File, Index: command.Index, TargetSize: textureTarget(command)}
}

func textureTarget(command drawCommand) uint32 {
	if len(command.Positions) != 8 || len(command.UV) != 8 {
		return 512
	}
	width := projectedTextureEdge(command, 2)
	height := projectedTextureEdge(command, 6)
	desired := min(512, max(4, width, height))
	return uint32(math.Exp2(math.Ceil(math.Log2(desired))))
}

func projectedTextureEdge(command drawCommand, corner int) float64 {
	screen := math.Hypot(command.Positions[corner]-command.Positions[0], command.Positions[corner+1]-command.Positions[1])
	uv := math.Hypot(command.UV[corner]-command.UV[0], command.UV[corner+1]-command.UV[1])
	if uv == 0 {
		return 0
	}
	return screen / uv
}

type textureTriangle struct {
	x, y, dx1, dy1, dx2, dy2, determinant float64
	u, v, du1, dv1, du2, dv2              float64
}

func sceneTriangle(command build.TreeDrawCommand, a, b, c int) textureTriangle {
	a, b, c = 2*a, 2*b, 2*c
	points, uv := command.Positions, command.UV
	x, y := points[a], points[a+1]
	dx1, dy1, dx2, dy2 := points[b]-x, points[b+1]-y, points[c]-x, points[c+1]-y
	u, v := uv[a], uv[a+1]
	return textureTriangle{x, y, dx1, dy1, dx2, dy2, dx1*dy2 - dx2*dy1,
		u, v, uv[b] - u, uv[b+1] - v, uv[c] - u, uv[c+1] - v}
}

func (t textureTriangle) sample(x, y float64) (float64, float64, bool) {
	if t.determinant == 0 {
		return 0, 0, false
	}
	dx, dy := x-t.x, y-t.y
	a := (dx*t.dy2 - dy*t.dx2) / t.determinant
	b := (t.dx1*dy - t.dy1*dx) / t.determinant
	inside := a >= 0 && b >= 0 && a+b <= 1
	return t.u + a*t.du1 + b*t.du2, t.v + a*t.dv1 + b*t.dv2, inside
}

func paintTreeCommand(frame *image.RGBA, texture *image.NRGBA, command build.TreeDrawCommand) {
	if len(command.Positions) != 8 || len(command.UV) != 8 || len(command.Color) != 4 {
		return
	}
	triangles := [2]textureTriangle{sceneTriangle(command, 0, 1, 2), sceneTriangle(command, 0, 2, 3)}
	origin := frame.Bounds().Min
	area := commandBox(command.Positions).Add(origin).Intersect(frame.Bounds())
	paintTreePixels(frame, texture, command, triangles, area)
}

func paintTreePixels(frame *image.RGBA, texture *image.NRGBA, command build.TreeDrawCommand, triangles [2]textureTriangle, area image.Rectangle) {
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			at := image.Pt(x, y).Sub(frame.Bounds().Min)
			u, v, inside := textureSample(triangles, float64(at.X)+0.5, float64(at.Y)+0.5)
			if inside {
				pixel := texturePixel(texture, u, v, command.Clamp)
				paintTreePixel(frame, x, y, pixel, command)
			}
		}
	}
}

func textureSample(triangles [2]textureTriangle, x, y float64) (float64, float64, bool) {
	u, v, inside := triangles[0].sample(x, y)
	if inside {
		return u, v, true
	}
	return triangles[1].sample(x, y)
}

func texturePixel(texture *image.NRGBA, u, v float64, clamp bool) color.NRGBA {
	if !clamp {
		u, v = u-math.Floor(u), v-math.Floor(v)
	}
	bounds := texture.Bounds()
	x := min(bounds.Dx()-1, max(0, int(u*float64(bounds.Dx()))))
	y := min(bounds.Dy()-1, max(0, int(v*float64(bounds.Dy()))))
	return texture.NRGBAAt(bounds.Min.X+x, bounds.Min.Y+y)
}

func paintTreePixel(frame *image.RGBA, x, y int, pixel color.NRGBA, command build.TreeDrawCommand) {
	alpha := float64(pixel.A) / 255 * command.Color[3]
	if alpha <= 0 {
		return
	}
	base := frame.RGBAAt(x, y)
	mix := 1 - alpha
	if command.Blend == "ADD" {
		mix = 1
	}
	blend := func(front, back uint8, color float64) uint8 {
		return uint8(min(255, max(0, math.Round(float64(front)*color*alpha+float64(back)*mix))))
	}
	frame.SetRGBA(x, y, color.RGBA{
		R: blend(pixel.R, base.R, command.Color[0]), G: blend(pixel.G, base.G, command.Color[1]),
		B: blend(pixel.B, base.B, command.Color[2]), A: 255,
	})
}

func commandBox(points []float64) image.Rectangle {
	left, top, right, bottom := points[0], points[1], points[0], points[1]
	for index := 2; index < len(points); index += 2 {
		x, y := points[index], points[index+1]
		left, top, right, bottom = min(left, x), min(top, y), max(right, x), max(bottom, y)
	}
	return image.Rect(int(math.Floor(left)), int(math.Floor(top)), int(math.Ceil(right)), int(math.Ceil(bottom)))
}
