package pob

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
	dds "github.com/kaiser-chris/dds-go"
	"github.com/klauspost/compress/zstd"
)

const (
	maxTextureFile   = 64 << 20
	maxTextureDDS    = 128 << 20
	maxTexturePNG    = 12 << 20
	maxTexturePixels = 16 << 20
)

// Последний атлас переиспользуется при загрузке следующих слоёв этого файла.
type treeTextureCache struct {
	mu   sync.Mutex
	name string
	dds  *dds.File
}

func (c *Calculator) markTreeAssets(scene *build.TreeScene) {
	if scene == nil {
		return
	}
	if c.installation.SHA256 != "" {
		scene.AssetID = c.installation.SHA256
		return
	}
	scene.AssetID = fmt.Sprintf("%x", sha256.Sum256([]byte(c.layout.Source)))
}

func (c *Calculator) ReadTreeTexture(ctx context.Context, query build.TreeTextureQuery) (build.TreeTexture, error) {
	if err := query.Validate(); err != nil {
		return build.TreeTexture{}, err
	}
	if err := ctx.Err(); err != nil {
		return build.TreeTexture{}, err
	}
	if query.File == "" {
		pixel := image.NewNRGBA(image.Rect(0, 0, 1, 1))
		pixel.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		return encodeTreePNG(pixel, 0)
	}
	if strings.HasSuffix(query.File, ".png") {
		return c.readTreePNG(query.File)
	}
	return c.readTreeDDS(query)
}

func (c *Calculator) readTreePNG(name string) (build.TreeTexture, error) {
	data, err := c.openTreeAsset(name, maxTexturePNG)
	if err != nil {
		return build.TreeTexture{}, err
	}
	info, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return build.TreeTexture{}, err
	}
	if !validTextureSize(info.Width, info.Height) {
		return build.TreeTexture{}, fmt.Errorf("изображение PoB превышает допустимый размер")
	}
	return build.TreeTexture{Width: uint32(info.Width), Height: uint32(info.Height), PNG: data}, nil
}

func (c *Calculator) readTreeDDS(query build.TreeTextureQuery) (build.TreeTexture, error) {
	atlas, err := c.treeDDS(query.File)
	if err != nil {
		return build.TreeTexture{}, err
	}
	level := textureMipLevel(atlas, int(max(1, query.TargetSize)))
	if !validTextureSize(max(1, atlas.Width>>level), max(1, atlas.Height>>level)) {
		return build.TreeTexture{}, fmt.Errorf("слой DDS превышает допустимый размер")
	}
	pixels, err := atlas.Decode(int(max(1, query.Index))-1, level)
	if err != nil {
		return build.TreeTexture{}, err
	}
	return encodeTreePNG(pixels, level)
}

func validTextureSize(width, height int) bool {
	return width > 0 && height > 0 && width <= 8192 && height <= 8192 && width*height <= maxTexturePixels
}

func textureMipLevel(atlas *dds.File, target int) int {
	selected := 0
	for level := range atlas.MipLevels {
		width, height := max(1, atlas.Width>>level), max(1, atlas.Height>>level)
		if min(width, height) < target {
			break
		}
		if !atlas.Format.Compressed() || width%4 == 0 && height%4 == 0 {
			selected = level
		}
	}
	return selected
}

func encodeTreePNG(pixels image.Image, level int) (build.TreeTexture, error) {
	var data bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := encoder.Encode(&data, pixels); err != nil {
		return build.TreeTexture{}, err
	}
	if data.Len() > maxTexturePNG {
		return build.TreeTexture{}, fmt.Errorf("PNG текстуры превышает допустимый размер")
	}
	size := pixels.Bounds().Size()
	return build.TreeTexture{Width: uint32(size.X), Height: uint32(size.Y), MipLevel: uint32(level), PNG: data.Bytes()}, nil
}

func (c *Calculator) treeDDS(name string) (*dds.File, error) {
	cache := &c.treeTextures
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.name == name && cache.dds != nil {
		return cache.dds, nil
	}
	compressed, err := c.openTreeAsset(name, maxTextureFile)
	if err != nil {
		return nil, err
	}
	reader, err := zstd.NewReader(bytes.NewReader(compressed), zstd.WithDecoderConcurrency(1), zstd.WithDecoderMaxMemory(maxTextureDDS))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	data, err := readTextureBytes(reader, maxTextureDDS)
	if err != nil {
		return nil, err
	}
	atlas, err := dds.Parse(data)
	if err != nil {
		return nil, err
	}
	if atlas.Dimension != dds.Texture2D || atlas.Images() > 4096 || atlas.MipLevels > 16 {
		return nil, fmt.Errorf("неподдерживаемые параметры атласа DDS")
	}
	cache.name, cache.dds = name, atlas
	return atlas, nil
}

func (c *Calculator) openTreeAsset(name string, limit int64) ([]byte, error) {
	root, err := os.OpenRoot(c.layout.Source)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	file, err := root.Open(filepath.FromSlash(name))
	if err != nil {
		return nil, fmt.Errorf("текстура недоступна в установке PoB: %q", name)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("недопустимый файл текстуры PoB: %q", name)
	}
	return readTextureBytes(file, limit)
}

func readTextureBytes(source io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(source, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("файл текстуры превышает допустимый размер")
	}
	return data, nil
}
