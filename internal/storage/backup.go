package storage

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
	bolt "go.etcd.io/bbolt"
)

type Verification struct {
	Path         string `json:"path"`
	Schema       string `json:"schema"`
	Records      uint64 `json:"records"`
	LogicalBytes uint64 `json:"logical_bytes"`
}

type legacyView struct{ *Journal }

func (s legacyView) Load(ctx context.Context, apply func(application.Change) error) error {
	return s.Journal.Load(ctx, func(change application.Change) error {
		if change.Apply != nil || change.Calculate != nil {
			return fmt.Errorf("новая команда в журнале версии 1")
		}
		return apply(change)
	})
}

func (s *Journal) migrateLegacy(ctx context.Context) error {
	if _, err := application.OpenService(ctx, legacyView{s}); err != nil {
		return err
	}
	backup := s.path + ".v1-" + rand.Text() + ".bak"
	if err := s.Backup(ctx, backup); err != nil {
		return fmt.Errorf("не создана резервная копия перед миграцией: %w", err)
	}
	err := s.db.Update(func(tx *bolt.Tx) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		var total uint64
		if err := tx.Bucket(changesBucket).ForEach(func(_, value []byte) error { total += uint64(len(value)); return nil }); err != nil {
			return err
		}
		counter := make([]byte, 8)
		binary.BigEndian.PutUint64(counter, total)
		meta := tx.Bucket(metadataBucket)
		if err := meta.Put(bytesKey, counter); err != nil {
			return err
		}
		return meta.Put(schemaKey, schemaVersion)
	})
	if err == nil {
		s.migrationBackup = backup
	}
	return err
}

func (s *Journal) MigrationBackupPath() string { return s.migrationBackup }

type contextWriter struct {
	ctx    context.Context
	writer io.Writer
}

func (w contextWriter) Write(data []byte) (int, error) {
	if err := w.ctx.Err(); err != nil {
		return 0, err
	}
	return w.writer.Write(data)
}

// Backup не перезаписывает существующий файл и подтверждает успех после Sync.
func (s *Journal) Backup(ctx context.Context, path string) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	info, err := output.Stat()
	if err != nil {
		output.Close()
		return err
	}
	success := false
	defer func() {
		output.Close()
		if !success {
			if current, statErr := os.Lstat(path); statErr == nil && os.SameFile(info, current) {
				_ = os.Remove(path)
			}
		}
	}()
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.db.View(func(tx *bolt.Tx) error { _, err := tx.WriteTo(contextWriter{ctx: ctx, writer: output}); return err }); err != nil {
		return err
	}
	if err := output.Sync(); err != nil {
		return err
	}
	if err := output.Close(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	success = true
	return nil
}

func openReadOnly(ctx context.Context, path string) (*Journal, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > 512<<20 {
		return nil, fmt.Errorf("неверный файл резервной копии")
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{ReadOnly: true, Timeout: time.Second})
	if err != nil {
		return nil, err
	}
	journal := &Journal{db: db, path: path}
	var version string
	err = db.View(func(tx *bolt.Tx) error {
		meta := tx.Bucket(metadataBucket)
		if meta == nil || tx.Bucket(changesBucket) == nil {
			return fmt.Errorf("нет схемы журнала")
		}
		version = string(meta.Get(schemaKey))
		if version != "1" && version != "2" {
			return fmt.Errorf("неподдерживаемая схема %s", version)
		}
		var first error
		for issue := range tx.Check() {
			if first == nil {
				first = issue
			}
		}
		return first
	})
	if err == nil {
		if version == "1" {
			_, err = application.OpenService(ctx, legacyView{journal})
		} else {
			_, err = application.OpenService(ctx, journal)
		}
	}
	if err != nil {
		db.Close()
		return nil, err
	}
	return journal, nil
}

func (s *Journal) verification() (Verification, error) {
	result := Verification{Path: s.path}
	err := s.db.View(func(tx *bolt.Tx) error {
		result.Schema = string(tx.Bucket(metadataBucket).Get(schemaKey))
		result.Records = tx.Bucket(changesBucket).Sequence()
		return tx.Bucket(changesBucket).ForEach(func(_, value []byte) error { result.LogicalBytes += uint64(len(value)); return nil })
	})
	return result, err
}

// Verify проверяет физическую структуру, суммы, порядок команд и восстановление состояния.
func Verify(ctx context.Context, path string) (Verification, error) {
	journal, err := openReadOnly(ctx, path)
	if err != nil {
		return Verification{}, err
	}
	defer journal.Close()
	return journal.verification()
}

// Restore создаёт новый профиль; имеющаяся база никогда не перезаписывается.
func Restore(ctx context.Context, source, base, profile string) (Verification, error) {
	target, err := Path(base, profile)
	if err != nil {
		return Verification{}, err
	}
	if _, err := os.Lstat(target); err == nil {
		return Verification{}, fmt.Errorf("база профиля уже существует")
	} else if !os.IsNotExist(err) {
		return Verification{}, err
	}
	journal, err := openReadOnly(ctx, source)
	if err != nil {
		return Verification{}, err
	}
	defer journal.Close()
	if err := privatefs.EnsureDir(filepath.Dir(target)); err != nil {
		return Verification{}, err
	}
	if err := journal.Backup(ctx, target); err != nil {
		return Verification{}, err
	}
	result, err := journal.verification()
	result.Path = target
	return result, err
}
