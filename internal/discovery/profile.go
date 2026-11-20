package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
)

var ErrLocked = errors.New("профиль занят другим процессом")

type Record struct {
	InstanceID      string `json:"instance_id"`
	Address         string `json:"address"`
	Token           string `json:"token"`
	ProtocolVersion string `json:"protocol_version"`
	Mode            string `json:"mode"`
	PID             int    `json:"pid"`
}

type Profile struct{ Dir, Name string }

// Open создаёт только собственные подкаталоги; базовый каталог не меняется.
func Open(base, name string) (Profile, error) {
	if err := privatefs.ValidateProfile(name); err != nil {
		return Profile{}, err
	}
	if base == "" {
		var err error
		base, err = os.UserCacheDir()
		if err != nil {
			return Profile{}, err
		}
	}
	root, err := filepath.Abs(filepath.Join(base, "pob2-engine"))
	if err != nil {
		return Profile{}, err
	}
	if err := privatefs.EnsureDir(root); err != nil {
		return Profile{}, err
	}
	dir := filepath.Join(root, name)
	if err := privatefs.EnsureDir(dir); err != nil {
		return Profile{}, err
	}
	return Profile{Dir: dir, Name: name}, nil
}

func (p Profile) RecordPath() string { return filepath.Join(p.Dir, "connection.json") }

func (p Profile) Read() (Record, error) {
	f, err := openRecord(p.RecordPath())
	if err != nil {
		return Record{}, err
	}
	defer f.Close()
	var record Record
	if err := json.NewDecoder(io.LimitReader(f, 65536)).Decode(&record); err != nil {
		return Record{}, err
	}
	if record.InstanceID == "" || record.Token == "" || record.Address == "" {
		return Record{}, fmt.Errorf("неполный файл подключения")
	}
	return record, nil
}

// Publish вызывается владельцем server.lock; замена файла выполняется атомарно.
func (p Profile) Publish(record Record) error {
	f, err := os.CreateTemp(p.Dir, "connection-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err := json.NewEncoder(f).Encode(record); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return replaceRecord(f.Name(), p.RecordPath())
}

// RemoveOwn вызывается до освобождения server.lock, поэтому чужой файл сохраняется.
func (p Profile) RemoveOwn(instanceID string) error {
	record, err := p.Read()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if record.InstanceID != instanceID {
		return nil
	}
	return os.Remove(p.RecordPath())
}

type Lock struct{ file *os.File }

func (l *Lock) Close() error { return l.file.Close() }

func (p Profile) TryLock(name string) (*Lock, error) {
	if name != "server" && name != "launch" {
		return nil, fmt.Errorf("неизвестная блокировка")
	}
	f, err := os.OpenFile(filepath.Join(p.Dir, name+".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := lockFile(f); err != nil {
		f.Close()
		return nil, err
	}
	return &Lock{file: f}, nil
}

func (p Profile) Lock(ctx context.Context, name string) (*Lock, error) {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		lock, err := p.TryLock(name)
		if !errors.Is(err, ErrLocked) {
			return lock, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
