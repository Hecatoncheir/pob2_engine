package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
	bolt "go.etcd.io/bbolt"
)

var (
	metadataBucket = []byte("metadata")
	changesBucket  = []byte("changes")
	schemaKey      = []byte("schema_version")
	schemaVersion  = []byte("1")
)

// Journal сохраняет журнал принятых команд транзакциями bbolt с включённым Sync.
type Journal struct {
	mu       sync.Mutex
	db       *bolt.DB
	path     string
	fault    error
	closed   bool
	closeErr error
}

func Path(base, profile string) (string, error) {
	if err := privatefs.ValidateProfile(profile); err != nil {
		return "", err
	}
	if base == "" {
		var err error
		base, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Abs(filepath.Join(base, "pob2-engine", "profiles", profile, "sessions.db"))
}

func Open(ctx context.Context, base, profile string) (*Journal, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := Path(base, profile)
	if err != nil {
		return nil, err
	}
	if err := privatefs.EnsureDir(filepath.Dir(path)); err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	created := errors.Is(err, os.ErrNotExist)
	if err != nil && !created {
		return nil, err
	}
	if !created && (!info.Mode().IsRegular() || info.Size() == 0) {
		return nil, fmt.Errorf("некорректный файл хранилища: %s", path)
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть хранилище %s: %w", path, err)
	}
	journal := &Journal{db: db, path: path}
	if err := journal.prepare(ctx, created); err != nil {
		db.Close()
		return nil, err
	}
	return journal, nil
}

func (s *Journal) Path() string { return s.path }

func (s *Journal) prepare(ctx context.Context, created bool) error {
	if created {
		if err := s.db.Update(func(tx *bolt.Tx) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if tx.Bucket(metadataBucket) != nil {
				return nil
			}
			if key, _ := tx.Cursor().First(); key != nil {
				return fmt.Errorf("неизвестный формат существующей базы")
			}
			meta, err := tx.CreateBucket(metadataBucket)
			if err != nil {
				return err
			}
			if err := meta.Put(schemaKey, schemaVersion); err != nil {
				return err
			}
			_, err = tx.CreateBucket(changesBucket)
			return err
		}); err != nil {
			return err
		}
	}
	return s.db.View(func(tx *bolt.Tx) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		meta := tx.Bucket(metadataBucket)
		if meta == nil || tx.Bucket(changesBucket) == nil {
			return fmt.Errorf("отсутствует схема хранилища сессий")
		}
		if !bytes.Equal(meta.Get(schemaKey), schemaVersion) {
			return fmt.Errorf("неподдерживаемая версия хранилища: %q", meta.Get(schemaKey))
		}
		var first error
		for err := range tx.Check() {
			if first == nil {
				first = err
			}
		}
		return first
	})
}

func (s *Journal) Load(ctx context.Context, apply func(application.Change) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(changesBucket)
		cursor := bucket.Cursor()
		sequence := uint64(0)
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			if err := ctx.Err(); err != nil {
				return err
			}
			sequence++
			if len(key) != 8 || binary.BigEndian.Uint64(key) != sequence {
				return fmt.Errorf("нарушен порядок команд хранилища")
			}
			change, err := decode(value)
			if err != nil {
				return fmt.Errorf("повреждена команда %d: %w", sequence, err)
			}
			if err := apply(change); err != nil {
				return fmt.Errorf("не удалось восстановить команду %d: %w", sequence, err)
			}
		}
		if sequence != bucket.Sequence() {
			return fmt.Errorf("не совпадает счётчик команд хранилища")
		}
		return ctx.Err()
	})
}

func encode(change application.Change) ([]byte, error) {
	payload, err := json.Marshal(change)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(payload)
	return append(digest[:], payload...), nil
}

func decode(value []byte) (application.Change, error) {
	var change application.Change
	if len(value) <= sha256.Size {
		return change, fmt.Errorf("неполная запись")
	}
	payload := value[sha256.Size:]
	digest := sha256.Sum256(payload)
	if !bytes.Equal(digest[:], value[:sha256.Size]) {
		return change, fmt.Errorf("не совпадает контрольная сумма")
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&change); err != nil {
		return change, err
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return change, fmt.Errorf("лишние данные записи")
	}
	return change, nil
}

func (s *Journal) Append(ctx context.Context, change application.Change) error {
	data, err := encode(change)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.fault != nil {
		return fmt.Errorf("запись заблокирована до перезапуска: %w", s.fault)
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		bucket := tx.Bucket(changesBucket)
		sequence, err := bucket.NextSequence()
		if err != nil {
			return err
		}
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, sequence)
		if bucket.Get(key) != nil {
			return fmt.Errorf("повтор номера команды хранилища")
		}
		if err := bucket.Put(key, data); err != nil {
			return err
		}
		return ctx.Err()
	})
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		s.fault = err
	}
	return err
}

func (s *Journal) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.closeErr = s.db.Close()
		s.closed = true
	}
	return s.closeErr
}
