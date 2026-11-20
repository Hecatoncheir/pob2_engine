package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	bolt "go.etcd.io/bbolt"
)

func TestBackupRestoreFullBuildAndOriginalRequestResult(t *testing.T) {
	j, service := openService(t, t.TempDir(), "main")
	_, err := service.Create(t.Context(), application.CreateSession{SessionID: "main", BuildName: "Сборка", RequestID: "create"})
	if err != nil {
		t.Fatal(err)
	}
	xml, err := os.ReadFile("../../testdata/reference/fireball.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Apply(t.Context(), application.ApplyBuild{SessionID: "main", ExpectedRevision: 1, RequestID: "import", XML: string(xml)})
	if err != nil {
		t.Fatal(err)
	}
	cmd := application.CalculateBuild{SessionID: "main", ExpectedRevision: 2, RequestID: "calculate"}
	saved, err := service.Calculate(t.Context(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(t.TempDir(), "backup.db")
	if err := j.Backup(t.Context(), backup); err != nil {
		t.Fatal(err)
	}
	verified, err := Verify(t.Context(), backup)
	if err != nil || verified.Schema != "2" || verified.Records != 3 {
		t.Fatalf("проверка: %+v %v", verified, err)
	}
	base := t.TempDir()
	if _, err := Restore(t.Context(), backup, base, "copy"); err != nil {
		t.Fatal(err)
	}
	_, restored := openService(t, base, "copy")
	if got, err := restored.Snapshot(t.Context(), "main"); err != nil || got != saved {
		t.Fatal("сборка или расчёт потеряны")
	}
	if got, err := restored.Calculate(t.Context(), cmd); err != nil || got != saved {
		t.Fatal("потерян первоначальный результат запроса")
	}
	if _, err := Restore(t.Context(), backup, base, "copy"); err == nil {
		t.Fatal("перезаписана существующая база")
	}
}

func legacyFile(t *testing.T, corrupt bool) (string, []byte) {
	t.Helper()
	base := t.TempDir()
	j, service := openService(t, base, "legacy")
	if _, err := service.Create(t.Context(), application.CreateSession{SessionID: "main", BuildName: "Старая", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	if err := j.db.Update(func(tx *bolt.Tx) error {
		meta := tx.Bucket(metadataBucket)
		if err := meta.Put(schemaKey, []byte("1")); err != nil {
			return err
		}
		if err := meta.Delete(bytesKey); err != nil {
			return err
		}
		if corrupt {
			bucket := tx.Bucket(changesBucket)
			key, value := bucket.Cursor().First()
			change, err := decode(value)
			if err != nil {
				return err
			}
			change.Result.Revision = 999
			encoded, err := encode(change)
			if err != nil {
				return err
			}
			return bucket.Put(key, encoded)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(j.path)
	if err != nil {
		t.Fatal(err)
	}
	return base, data
}

func TestLegacyMigrationCreatesVerifiedBackup(t *testing.T) {
	base, _ := legacyFile(t, false)
	j, service := openService(t, base, "legacy")
	if j.MigrationBackupPath() == "" {
		t.Fatal("миграция выполнена без резервной копии")
	}
	backup, err := Verify(t.Context(), j.MigrationBackupPath())
	if err != nil || backup.Schema != "1" || backup.Records != 1 {
		t.Fatalf("неверная копия старой схемы: %+v %v", backup, err)
	}
	if got, err := service.Snapshot(t.Context(), "main"); err != nil || got.BuildName != "Старая" {
		t.Fatal("потеряно состояние при миграции")
	}
	if _, err := service.Rename(t.Context(), application.RenameBuild{SessionID: "main", BuildName: "Новая", ExpectedRevision: 1, RequestID: "rename"}); err != nil {
		t.Fatal(err)
	}
}

func TestCorruptLegacyIsNotMigratedOrModified(t *testing.T) {
	base, before := legacyFile(t, true)
	j, err := Open(t.Context(), base, "legacy")
	if err == nil {
		j.Close()
		t.Fatal("повреждённая старая база мигрирована")
	}
	path, err := Path(base, "legacy")
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("исходная база изменена")
	}
	backups, err := filepath.Glob(path + ".v1-*.bak")
	if err != nil || len(backups) != 0 {
		t.Fatal("создана копия непроверенного состояния")
	}
}

func TestCanceledAndExistingBackupDoNotDestroyFiles(t *testing.T) {
	j, _ := openService(t, t.TempDir(), "main")
	path := filepath.Join(t.TempDir(), "backup.db")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := j.Backup(ctx, path); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("отмена оставила файл")
	}
	if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := j.Backup(t.Context(), path); err == nil {
		t.Fatal("существующая копия перезаписана")
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, []byte("keep")) {
		t.Fatal("посторонний файл изменён")
	}
	if _, err := Restore(t.Context(), path, t.TempDir(), "invalid"); err == nil {
		t.Fatal("повреждённая копия восстановлена")
	}
}

func TestJournalCapacityFailureDoesNotPoisonWriter(t *testing.T) {
	j, service := openService(t, t.TempDir(), "main")
	initial, err := service.Create(t.Context(), application.CreateSession{SessionID: "main", BuildName: "До", RequestID: "create"})
	if err != nil {
		t.Fatal(err)
	}
	var original []byte
	if err := j.db.Update(func(tx *bolt.Tx) error {
		meta := tx.Bucket(metadataBucket)
		original = bytes.Clone(meta.Get(bytesKey))
		full := make([]byte, 8)
		binary.BigEndian.PutUint64(full, maxJournalBytes)
		return meta.Put(bytesKey, full)
	}); err != nil {
		t.Fatal(err)
	}
	cmd := application.RenameBuild{SessionID: "main", BuildName: "После", ExpectedRevision: 1, RequestID: "rename"}
	_, err = service.Rename(t.Context(), cmd)
	var failure *application.Error
	if !errors.As(err, &failure) || failure.Code != application.ResourceExhausted || j.fault != nil {
		t.Fatalf("лимит принят за ошибку диска: %v", err)
	}
	if got, _ := service.Snapshot(t.Context(), "main"); got != initial {
		t.Fatal("неподтверждённая правка опубликована")
	}
	if err := j.db.Update(func(tx *bolt.Tx) error { return tx.Bucket(metadataBucket).Put(bytesKey, original) }); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Rename(t.Context(), cmd); err != nil {
		t.Fatal(err)
	}
}
