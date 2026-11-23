package build

import (
	"fmt"
	"path"
	"strings"
)

type TreeTextureQuery struct {
	File       string `json:"file"`
	Index      uint32 `json:"index,omitempty"`
	TargetSize uint32 `json:"target_size,omitempty"`
}

func (q TreeTextureQuery) Validate() error {
	if q.Index > 4096 || q.TargetSize > 8192 {
		return fmt.Errorf("слой или размер текстуры вне диапазона")
	}
	if q.File == "" {
		return nil
	}
	if q.File != path.Clean(q.File) || strings.ContainsAny(q.File, "\\:\x00") || len(q.File) > 512 {
		return fmt.Errorf("недопустимый путь текстуры PoB")
	}
	if !strings.HasPrefix(q.File, "Assets/") && !strings.HasPrefix(q.File, "TreeData/") {
		return fmt.Errorf("текстура должна находиться в Assets или TreeData")
	}
	if !strings.HasSuffix(q.File, ".png") && !strings.HasSuffix(q.File, ".dds.zst") {
		return fmt.Errorf("формат текстуры: PNG или DDS Zstandard")
	}
	return nil
}

type TreeTexture struct {
	Width    uint32 `json:"width"`
	Height   uint32 `json:"height"`
	MipLevel uint32 `json:"mip_level"`
	PNG      []byte `json:"png"`
}
