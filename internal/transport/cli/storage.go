package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/storage"
)

func (r runner) storage(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return &UsageError{Err: fmt.Errorf("нужна операция storage backup|verify|restore")}
	}
	operation := args[0]
	if operation != "backup" && operation != "verify" && operation != "restore" {
		return &UsageError{Err: fmt.Errorf("неизвестная операция storage %q", operation)}
	}
	fs := r.flags("storage " + operation)
	profile := fs.String("profile", "default", "профиль остановленного сервера или новый профиль восстановления")
	directory := fs.String("data-dir", "", "базовый каталог данных")
	file := fs.String("file", "", "файл резервной копии")
	if err := parse(fs, args[1:]); err != nil {
		return err
	}
	if *file == "" {
		return &UsageError{Err: fmt.Errorf("нужен --file")}
	}
	if operation == "verify" {
		result, err := storage.Verify(ctx, *file)
		if err != nil {
			return err
		}
		return r.write(result)
	}
	if operation == "restore" {
		result, err := storage.Restore(ctx, *file, *directory, *profile)
		if err != nil {
			return err
		}
		return r.write(result)
	}
	path, err := storage.Path(*directory, *profile)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	journal, err := storage.Open(ctx, *directory, *profile)
	if err != nil {
		return err
	}
	defer journal.Close()
	if _, err := application.OpenService(ctx, journal); err != nil {
		return err
	}
	if err := journal.Backup(ctx, *file); err != nil {
		return err
	}
	result, err := storage.Verify(ctx, *file)
	if err != nil {
		return err
	}
	return r.write(result)
}
