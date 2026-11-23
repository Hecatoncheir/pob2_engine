package cli

import (
	"context"
	"fmt"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

func (r runner) treeTexture(ctx context.Context, args []string) error {
	fs := r.flags("tree-texture")
	connection := connectFlagsWithTimeout(fs, buildTimeout)
	file := fs.String("file", "", "ссылка из команды отрисовки")
	index := fs.Uint("index", 1, "слой DDS, начиная с единицы")
	target := fs.Uint("target-size", 1, "целевой размер текстуры")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *index > 4096 || *target > 8192 {
		return &UsageError{Err: fmt.Errorf("слой или размер текстуры вне диапазона")}
	}
	ctx, cancel := context.WithTimeout(ctx, connection.timeout)
	defer cancel()
	client, err := connection.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	query := build.TreeTextureQuery{File: *file, Index: uint32(*index), TargetSize: uint32(*target)}
	texture, err := client.TreeTexture(ctx, query)
	if err != nil {
		return err
	}
	return r.write(texture)
}
