package storage

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
	"github.com/Hecatoncheir/pob2_engine/internal/testreference"
	bolt "go.etcd.io/bbolt"
)

func openService(t *testing.T, base, profile string) (*Journal, *application.Service) {
	t.Helper()
	j, err := Open(t.Context(), base, profile)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = j.Close() })
	data, err := gamedata.Embedded()
	if err != nil {
		t.Fatal(err)
	}
	calculator, err := testreference.New(data)
	if err != nil {
		t.Fatal(err)
	}
	s, err := application.OpenServiceWithCalculator(t.Context(), j, calculator)
	if err != nil {
		t.Fatal(err)
	}
	return j, s
}

func TestRestartRestoresRevisionsAndReplayResults(t *testing.T) {
	base := t.TempDir()
	j, service := openService(t, base, "main")
	create := application.CreateSession{SessionID: "main", BuildName: "Начальная", RequestID: "create"}
	initial, err := service.Create(t.Context(), create)
	if err != nil {
		t.Fatal(err)
	}
	first := application.RenameBuild{SessionID: "main", BuildName: "Первая", ExpectedRevision: 1, RequestID: "first"}
	changed, err := service.Rename(t.Context(), first)
	if err != nil {
		t.Fatal(err)
	}
	last, err := service.Rename(t.Context(), application.RenameBuild{SessionID: "main", BuildName: "Последняя", ExpectedRevision: 2, RequestID: "last"})
	if err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	_, restored := openService(t, base, "main")
	if got, err := restored.Create(t.Context(), create); err != nil || got != initial {
		t.Fatalf("повтор создания: %v, %v", got, err)
	}
	if got, err := restored.Rename(t.Context(), first); err != nil || got != changed {
		t.Fatalf("повтор старой правки: %v, %v", got, err)
	}
	if got, err := restored.Snapshot(t.Context(), "main"); err != nil || got != last {
		t.Fatalf("текущий снимок: %v, %v", got, err)
	}
	first.BuildName = "Подмена"
	if _, err := restored.Rename(t.Context(), first); err == nil {
		t.Fatal("разрешена подмена записанной команды")
	}
}

func TestConcurrentSessionsRemainConsistentAfterRestart(t *testing.T) {
	base := t.TempDir()
	j, service := openService(t, base, "concurrent")
	var group sync.WaitGroup
	results := make(chan error, 8)
	for i := range 8 {
		group.Go(func() {
			id := fmt.Sprintf("session-%d", i)
			if _, err := service.Create(t.Context(), application.CreateSession{SessionID: id, BuildName: id, RequestID: "create-" + id}); err != nil {
				results <- err
				return
			}
			for revision := uint64(1); revision <= 8; revision++ {
				if _, err := service.Rename(t.Context(), application.RenameBuild{SessionID: id, BuildName: fmt.Sprint(revision), ExpectedRevision: revision, RequestID: fmt.Sprintf("rename-%d", revision)}); err != nil {
					results <- err
					return
				}
			}
			results <- nil
		})
	}
	group.Wait()
	for range 8 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	j.Close()
	_, restored := openService(t, base, "concurrent")
	for i := range 8 {
		got, err := restored.Snapshot(t.Context(), fmt.Sprintf("session-%d", i))
		if err != nil || got.Revision != 9 || got.BuildName != "8" {
			t.Fatalf("параллельная запись потеряна: %v, %v", got, err)
		}
	}
}

func TestUnknownSchemaAndCorruptionAreNotReset(t *testing.T) {
	for _, mode := range []string{"schema", "checksum", "sequence", "invalid-state"} {
		t.Run(mode, func(t *testing.T) {
			base := t.TempDir()
			j, service := openService(t, base, "main")
			if _, err := service.Create(t.Context(), application.CreateSession{SessionID: "main", BuildName: "Сохранённая", RequestID: "create"}); err != nil {
				t.Fatal(err)
			}
			if err := j.db.Update(func(tx *bolt.Tx) error {
				bucket := tx.Bucket(changesBucket)
				key := make([]byte, 8)
				binary.BigEndian.PutUint64(key, 1)
				switch mode {
				case "schema":
					return tx.Bucket(metadataBucket).Put(schemaKey, []byte("999"))
				case "checksum":
					data := append([]byte(nil), bucket.Get(key)...)
					data[len(data)-1] ^= 1
					return bucket.Put(key, data)
				case "sequence":
					return bucket.Delete(key)
				default:
					change, err := decode(bucket.Get(key))
					if err != nil {
						return err
					}
					change.Result.Revision = 999
					data, err := encode(change)
					if err != nil {
						return err
					}
					return bucket.Put(key, data)
				}
			}); err != nil {
				t.Fatal(err)
			}
			j.Close()
			before, err := os.ReadFile(j.Path())
			if err != nil {
				t.Fatal(err)
			}
			reopened, err := Open(t.Context(), base, "main")
			if err == nil {
				_, err = application.OpenService(t.Context(), reopened)
				reopened.Close()
			}
			if err == nil {
				t.Fatal("повреждённое хранилище принято")
			}
			after, err := os.ReadFile(j.Path())
			if err != nil {
				t.Fatal(err)
			}
			if sha256.Sum256(before) != sha256.Sum256(after) {
				t.Fatal("повреждённое хранилище изменено при запуске")
			}
		})
	}
}

func TestZeroOrForeignFileIsNotInitialized(t *testing.T) {
	for _, data := range [][]byte{nil, []byte("не база данных")} {
		base := t.TempDir()
		path, err := Path(base, "main")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		j, err := Open(t.Context(), base, "main")
		if err == nil {
			j.Close()
			t.Fatal("посторонний или пустой файл принят")
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != string(data) {
			t.Fatal("исходный файл перезаписан")
		}
	}
}

func TestCanceledWriteDoesNotAdvanceJournal(t *testing.T) {
	j, service := openService(t, t.TempDir(), "main")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := j.Append(ctx, application.Change{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("отмена: %v", err)
	}
	if _, err := service.Create(t.Context(), application.CreateSession{SessionID: "main", BuildName: "Сборка", RequestID: "create"}); err != nil {
		t.Fatal(err)
	}
	count := 0
	if err := j.Load(t.Context(), func(application.Change) error { count++; return nil }); err != nil || count != 1 {
		t.Fatalf("отменённая запись сохранена: %d, %v", count, err)
	}
}

func TestDatabaseWriteFailureBlocksFurtherWritesUntilRestart(t *testing.T) {
	base := t.TempDir()
	j, service := openService(t, base, "main")
	initial, err := service.Create(t.Context(), application.CreateSession{SessionID: "main", BuildName: "Сохранённая", RequestID: "create"})
	if err != nil {
		t.Fatal(err)
	}
	if err := j.db.Close(); err != nil {
		t.Fatal(err)
	}
	command := application.RenameBuild{SessionID: "main", BuildName: "После перезапуска", ExpectedRevision: 1, RequestID: "rename"}
	if _, err := service.Rename(t.Context(), command); err == nil {
		t.Fatal("ошибка базы не сообщена")
	}
	if j.fault == nil {
		t.Fatal("ошибка записи не заблокировала журнал")
	}
	if got, err := service.Snapshot(t.Context(), "main"); err != nil || got != initial {
		t.Fatal("неподтверждённое состояние опубликовано")
	}
	j.Close()
	_, restored := openService(t, base, "main")
	if got, err := restored.Rename(t.Context(), command); err != nil || got.Revision != 2 {
		t.Fatalf("восстановление после ошибки: %v, %v", got, err)
	}
}
