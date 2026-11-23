package mcpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type imageOptions struct {
	view, query string
	page, id    int
	width       int
	height      int
	labels      bool
}

func workbenchImageOptions(arguments map[string]any) (imageOptions, error) {
	view, _ := arguments["view"].(string)
	if view == "" {
		view = "overview"
	}
	query, _ := arguments["query"].(string)
	labels, _ := arguments["showLabels"].(bool)
	options := imageOptions{view: view, query: query, labels: labels,
		page: number(arguments["page"], 1), id: number(arguments["id"], 0),
		width: number(arguments["width"], 1600), height: number(arguments["height"], 1100)}
	if (view == "node" || view == "item") && options.id == 0 {
		return options, fmt.Errorf("для раздела %s требуется ID", view)
	}
	if view == "tree" && options.page != 1 {
		return options, fmt.Errorf("дерево помещается на одной странице")
	}
	return options, nil
}

func (s *workbenchServer) renderImage(ctx context.Context, arguments map[string]any) *mcp.CallToolResult {
	options, err := workbenchImageOptions(arguments)
	if err != nil {
		return workbenchFailure("INVALID_ARGUMENTS", err)
	}
	reply, err := s.session.Call(ctx, "getBuildImageData", arguments)
	if err != nil {
		return workbenchFailure("CALCULATION_FAILED", err)
	}
	var data imageData
	if err := json.Unmarshal(reply.Result, &data); err != nil {
		return workbenchFailure("INVALID_RESPONSE", err)
	}
	canvas, err := newWorkbenchCanvas(options, data)
	if err != nil {
		return workbenchFailure("INVALID_RESPONSE", err)
	}
	defer canvas.close()
	details, err := s.paintImage(ctx, canvas, options, data)
	if err != nil {
		return workbenchFailure("INVALID_ARGUMENTS", err)
	}
	return imageToolResult(canvas, reply.Revision, details)
}

func (s *workbenchServer) paintImage(ctx context.Context, canvas *workbenchCanvas, options imageOptions, data imageData) (map[string]any, error) {
	var tree map[string]any
	pages := 1
	var err error
	if options.view == "tree" {
		tree, err = s.assets.paintTree(ctx, canvas, options, data)
	} else {
		pages, err = canvas.paintRows(options, data)
	}
	if err != nil {
		return nil, err
	}
	warnings := len(data.warnings())
	canvas.box(image.Rect(28, options.height-44, options.width-28, options.height-43), imagePanel)
	canvas.text(32, options.height-18, fmt.Sprintf("Расчёт PoB · Предупреждений: %d · Страница %d/%d", warnings, options.page, pages), 14, imageMuted)
	if tree == nil {
		tree = map[string]any{}
	}
	return map[string]any{
		"view": options.view, "page": options.page, "pages": pages, "width": options.width, "height": options.height,
		"mimeType": "image/png", "warningCount": warnings, "buildInfo": data.info(), "tree": tree,
	}, nil
}

func imageToolResult(canvas *workbenchCanvas, revision int, details map[string]any) *mcp.CallToolResult {
	var image bytes.Buffer
	if err := png.Encode(&image, canvas.pixels); err != nil {
		return workbenchFailure("INVALID_RESPONSE", err)
	}
	if image.Len() > 4<<20 {
		return workbenchFailure("IMAGE_TOO_LARGE", fmt.Errorf("PNG превышает 4 МиБ; уменьшите размер изображения"))
	}
	result, err := workbenchResult(revision, details)
	if err != nil {
		return workbenchFailure("INVALID_RESPONSE", err)
	}
	text, err := json.Marshal(result)
	if err != nil {
		return workbenchFailure("INVALID_RESPONSE", err)
	}
	return &mcp.CallToolResult{StructuredContent: result,
		Content: []mcp.Content{&mcp.TextContent{Text: string(text)}, &mcp.ImageContent{MIMEType: "image/png", Data: image.Bytes()}}}
}

var (
	imageBackground = color.NRGBA{17, 22, 30, 255}
	imagePanel      = color.NRGBA{31, 40, 51, 255}
	imageText       = color.NRGBA{225, 231, 238, 255}
	imageMuted      = color.NRGBA{145, 163, 181, 255}
	imageGold       = color.NRGBA{235, 188, 103, 255}
	imageWarning    = color.NRGBA{249, 146, 131, 255}
	imageFont       = sync.OnceValues(func() (*opentype.Font, error) { return opentype.Parse(goregular.TTF) })
)

type workbenchCanvas struct {
	pixels *image.RGBA
	faces  map[int]font.Face
}

func newWorkbenchCanvas(options imageOptions, data imageData) (*workbenchCanvas, error) {
	parsed, err := imageFont()
	if err != nil {
		return nil, err
	}
	canvas := &workbenchCanvas{pixels: image.NewRGBA(image.Rect(0, 0, options.width, options.height)), faces: map[int]font.Face{}}
	for _, size := range []int{13, 14, 16, 17, 18, 19, 27} {
		face, err := opentype.NewFace(parsed, &opentype.FaceOptions{Size: float64(size), DPI: 72, Hinting: font.HintingFull})
		if err != nil {
			canvas.close()
			return nil, err
		}
		canvas.faces[size] = face
	}
	canvas.drawHeader(options, data)
	return canvas, nil
}

func (c *workbenchCanvas) close() {
	for _, face := range c.faces {
		_ = face.Close()
	}
}

func (c *workbenchCanvas) drawHeader(options imageOptions, data imageData) {
	info := data.info()
	c.box(c.pixels.Bounds(), imageBackground)
	c.box(image.Rect(0, 0, options.width, 102), imagePanel)
	c.box(image.Rect(0, 0, 6, 102), imageGold)
	name := c.wrap(display(info["name"]), 27, options.width-64)
	c.text(32, 40, name[0], 27, imageText)
	subtitle := fmt.Sprintf("%s / %s · Уровень %s · %s", display(info["class"]), display(info["ascendancy"]),
		display(info["level"]), imageTitles[options.view])
	for offset, line := range c.wrap(subtitle, 17, options.width-64) {
		if offset == 2 {
			break
		}
		c.text(32, 68+offset*20, line, 17, imageMuted)
	}
}

func (c *workbenchCanvas) box(area image.Rectangle, fill color.Color) {
	draw.Draw(c.pixels, area, image.NewUniform(fill), image.Point{}, draw.Src)
}

func (c *workbenchCanvas) text(x, y int, text string, size int, fill color.Color) {
	drawing := font.Drawer{Dst: c.pixels, Src: image.NewUniform(fill), Face: c.faces[size], Dot: fixed.P(x, y)}
	drawing.DrawString(text)
}

func (c *workbenchCanvas) paintRows(options imageOptions, data imageData) (int, error) {
	rows, err := data.rows(options)
	if err != nil {
		return 0, err
	}
	visible := []imageRow{}
	for _, row := range rows {
		size, _ := imageStyle(row.style)
		for _, line := range c.wrap(row.text, size, options.width-100) {
			visible = append(visible, imageRow{line, row.style})
		}
	}
	count := max(1, (options.height-172)/27)
	pages := max(1, (len(visible)+count-1)/count)
	if options.page > pages {
		return 0, fmt.Errorf("страница %d недоступна; всего страниц: %d", options.page, pages)
	}
	start := (options.page - 1) * count
	for offset, row := range visible[start:min(len(visible), start+count)] {
		y := 118 + offset*27
		if row.style == "heading" {
			c.box(image.Rect(24, y, options.width-24, y+26), imagePanel)
		}
		size, fill := imageStyle(row.style)
		c.text(36, y+20, row.text, size, fill)
	}
	return pages, nil
}

func imageStyle(style string) (int, color.Color) {
	switch style {
	case "heading":
		return 19, imageGold
	case "muted":
		return 16, imageMuted
	case "warning":
		return 17, imageWarning
	default:
		return 18, imageText
	}
}

func (c *workbenchCanvas) wrap(text string, size, width int) []string {
	var lines []string
	for line := range strings.SplitSeq(text, "\n") {
		lines = append(lines, wrapImageLine(c.faces[size], line, width)...)
	}
	return lines
}

func wrapImageLine(face font.Face, line string, width int) []string {
	text := []rune(line)
	result := []string{}
	for len(text) != 0 {
		end := fittingRunes(face, text, width)
		result = append(result, strings.TrimRight(string(text[:end]), " "))
		text = text[end:]
		for len(text) != 0 && text[0] == ' ' {
			text = text[1:]
		}
	}
	if len(result) == 0 {
		return []string{""}
	}
	return result
}

func fittingRunes(face font.Face, line []rune, width int) int {
	length, gap := fixed.Int26_6(0), 0
	for index, character := range line {
		advance, _ := face.GlyphAdvance(character)
		length += advance
		if length.Ceil() > width && index > 0 {
			if gap > 0 {
				return gap
			}
			return index
		}
		if character == ' ' {
			gap = index + 1
		}
	}
	return len(line)
}
