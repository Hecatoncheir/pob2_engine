package pob

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func installForRollback(t *testing.T) (*Manager, release, release, []byte) {
	t.Helper()
	old, oldArchive := fixtureRelease(t, "v0.23.0")
	next, nextArchive := fixtureRelease(t, "v0.23.1")
	m := &Manager{root: t.TempDir(), client: &http.Client{Transport: archiveTransport(oldArchive)},
		probe: func(_ context.Context, i Installation) error {
			_, err := i.Layout()
			return err
		}}
	if _, err := m.installUpdate(t.Context(), old, State{}); err != nil {
		t.Fatal(err)
	}
	m.client.Transport = availableRelease(t, next, nextArchive)
	update, err := m.Update(t.Context(), true)
	if err != nil || update.InstalledVersion != next.Tag || update.PreviousVersion != old.Tag {
		t.Fatalf("предыдущая версия не записана: %+v, %v", update, err)
	}
	m.client.Transport = installTransport(func(*http.Request) (*http.Response, error) {
		t.Fatal("откат обратился к сети")
		return nil, errors.New("сеть недоступна")
	})
	return m, old, next, nextArchive
}

func TestPoBRollbackOfflineWithBrokenCurrentInstallation(t *testing.T) {
	m, old, next, _ := installForRollback(t)
	check, err := m.Rollback(t.Context(), false)
	if err != nil || check.Status != "rollback_available" || check.InstalledVersion != next.Tag {
		t.Fatalf("не обнаружена доступная установка для отката: %+v, %v", check, err)
	}
	working, err := m.Installed()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(working.Directory, "src", "Launch.lua")); err != nil {
		t.Fatal(err)
	}
	rolledBack, err := m.Rollback(t.Context(), true)
	if err != nil || rolledBack.Status != "rolled_back" || rolledBack.InstalledVersion != old.Tag || !rolledBack.RestartRequired {
		t.Fatalf("откат не активировал предыдущую установку: %+v, %v", rolledBack, err)
	}
	reopened := &Manager{root: m.root}
	requireActiveVersion(t, reopened, old.Tag)
	repeated, err := reopened.Rollback(t.Context(), true)
	if err != nil || repeated.Status != "rollback_unavailable" || repeated.InstalledVersion != old.Tag {
		t.Fatalf("повтор отката изменил установку: %+v, %v", repeated, err)
	}
}

func TestPoBRollbackCanceledBeforeCommitKeepsBothVersions(t *testing.T) {
	m, old, next, _ := installForRollback(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	m.probe = func(context.Context, Installation) error {
		cancel()
		return nil
	}
	if _, err := m.Rollback(ctx, true); !errors.Is(err, context.Canceled) {
		t.Fatalf("отмена отката не была возвращена: %v", err)
	}
	reopened := &Manager{root: m.root}
	requireActiveVersion(t, reopened, next.Tag)
	check, err := reopened.Rollback(t.Context(), false)
	if err != nil || check.Status != "rollback_available" || check.PreviousVersion != old.Tag {
		t.Fatalf("отмена затронула сохранённую версию: %+v, %v", check, err)
	}
}

func TestRejectCachedCandidateBeforeActivation(t *testing.T) {
	m, old, next, archive := installForRollback(t)
	if _, err := m.Rollback(t.Context(), true); err != nil {
		t.Fatal(err)
	}
	// Используем один и тот же набор байтов для воспроизводимого имени каталога.
	cached := m.locate(Installation{Version: next.Tag, SHA256: next.Assets[0].sha256()})
	if err := os.Remove(filepath.Join(cached.Directory, "src", "Launch.lua")); err != nil {
		t.Fatal(err)
	}
	m.client.Transport = availableRelease(t, next, archive)
	if _, err := m.Update(t.Context(), true); err == nil || !strings.Contains(err.Error(), "PoB в кэше") {
		t.Fatalf("не обнаружена повреждённая копия в кэше: %v", err)
	}
	requireActiveVersion(t, m, old.Tag)
}
