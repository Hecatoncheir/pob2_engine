package pob

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func fixtureRelease(t *testing.T, tag string) (release, []byte) {
	t.Helper()
	library, extension := luaLibrary("runtime", "runtime")
	pkg := tinyArchive(t, map[string]string{
		"src/Launch.lua": "return true", "runtime/lua/dkjson.lua": "return {}",
		filepath.ToSlash(library): "LuaJIT", "runtime/lua-utf8." + extension: "UTF-8",
	})
	return release{Tag: tag, commit: strings.Repeat("1", 40), Assets: []asset{{
		Name: portableName, Size: int64(len(pkg)), Digest: fmt.Sprintf("sha256:%x", sha256.Sum256(pkg)),
	}}}, pkg
}

func requireActiveVersion(t *testing.T, m *Manager, tag string) {
	t.Helper()
	active, err := m.Installed()
	if err != nil || active.Version != tag {
		t.Fatalf("ожидалась активная версия %s: %+v, %v", tag, active, err)
	}
	if _, err := sourceRoot(active.Directory); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentInstallPublishesOnePoB(t *testing.T) {
	rel, archive := fixtureRelease(t, "v0.23.1")
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	var downloads atomic.Int32
	started, finishProbe := make(chan struct{}, 1), make(chan struct{})
	m := Manager{root: t.TempDir(),
		client: &http.Client{Transport: installTransport(func(req *http.Request) (*http.Response, error) {
			downloads.Add(1)
			return archiveTransport(archive).RoundTrip(req)
		})},
		probe: func(ctx context.Context, _ Installation) error {
			started <- struct{}{}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-finishProbe:
				return nil
			}
		},
	}
	first := installAsync(ctx, &m, rel)
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("первая установка не дошла до запуска PoB")
	}
	if _, err := m.Installed(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("установка стала активной до пробного запуска: %v", err)
	}
	verifyInstallWait(t, &m, rel)
	second := installAsync(ctx, &m, rel)
	close(finishProbe)
	statuses := []string{installStatus(t, first), installStatus(t, second)}
	sort.Strings(statuses)
	if statuses[0] != "up_to_date" || statuses[1] != "updated" || downloads.Load() != 1 {
		t.Fatalf("ожидалась одна установка, статусы %v, загрузок %d", statuses, downloads.Load())
	}
	requireActiveVersion(t, &m, rel.Tag)
}

func verifyInstallWait(t *testing.T, m *Manager, rel release) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 150*time.Millisecond)
	defer cancel()
	if _, err := m.installUpdate(ctx, rel, State{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("вторая установка должна ждать блокировку до таймаута: %v", err)
	}
}

type installResult struct {
	state State
	err   error
}

func installAsync(ctx context.Context, m *Manager, rel release) <-chan installResult {
	result := make(chan installResult, 1)
	go func() {
		state, err := m.installUpdate(ctx, rel, State{})
		result <- installResult{state: state, err: err}
	}()
	return result
}

func installStatus(t *testing.T, result <-chan installResult) string {
	t.Helper()
	value := <-result
	if value.err != nil {
		t.Fatal(value.err)
	}
	return value.state.Status
}

func TestUpdateActivatesPoBForNextStart(t *testing.T) {
	old, oldArchive := fixtureRelease(t, "v0.23.1")
	next, nextArchive := fixtureRelease(t, "v0.23.2")
	m := Manager{root: t.TempDir(), client: &http.Client{Transport: archiveTransport(oldArchive)},
		probe: func(_ context.Context, installed Installation) error {
			_, err := installed.Layout()
			return err
		}}
	if _, err := m.installUpdate(t.Context(), old, State{}); err != nil {
		t.Fatal(err)
	}
	m.client.Transport = availableRelease(t, next, nextArchive)
	check, err := m.Update(t.Context(), false)
	if err != nil || check.Status != "available" || check.InstalledVersion != old.Tag || check.LatestVersion != next.Tag {
		t.Fatalf("обновление PoB не обнаружено: %+v, %v", check, err)
	}
	requireActiveVersion(t, &m, old.Tag)
	update, err := m.Update(t.Context(), true)
	if err != nil || update.Status != "updated" || update.InstalledVersion != next.Tag || !update.RestartRequired {
		t.Fatalf("новая версия PoB не активирована: %+v, %v", update, err)
	}
	reopened := Manager{root: m.root}
	requireActiveVersion(t, &reopened, next.Tag)
	if _, err := m.locate(Installation{Version: old.Tag, SHA256: old.Assets[0].sha256()}).Layout(); err != nil {
		t.Fatalf("предыдущая версия PoB недоступна: %v", err)
	}
}

func availableRelease(t *testing.T, next release, archive []byte) http.RoundTripper {
	t.Helper()
	metadata, err := json.Marshal(next)
	if err != nil {
		t.Fatal(err)
	}
	return installTransport(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case latestURL:
			return archiveTransport(metadata).RoundTrip(req)
		case "https://api.github.com/repos/" + repository + "/commits/" + next.Tag:
			return archiveTransport(fmt.Sprintf(`{"sha":%q}`, next.commit)).RoundTrip(req)
		case downloadURL(repository, next.Tag, portableName):
			return archiveTransport(archive).RoundTrip(req)
		default:
			return nil, fmt.Errorf("неожиданный URL: %s", req.URL)
		}
	})
}

func TestFailedInstallPreservesActivePoB(t *testing.T) {
	for _, failure := range []string{"download", "probe", "cancel_before_activation"} {
		t.Run(failure, func(t *testing.T) {
			verifyFailedInstall(t, failure)
		})
	}
}

func verifyFailedInstall(t *testing.T, failure string) {
	old, oldArchive := fixtureRelease(t, "v0.23.1")
	next, nextArchive := fixtureRelease(t, "v0.23.2")
	m := Manager{root: t.TempDir(), client: &http.Client{Transport: archiveTransport(oldArchive)},
		probe: func(context.Context, Installation) error { return nil }}
	if _, err := m.installUpdate(t.Context(), old, State{}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	m.client.Transport, m.probe = failedInstall(failure, nextArchive, cancel)
	if _, err := m.installUpdate(ctx, next, State{}); err == nil {
		t.Fatal("неудачная установка завершилась без ошибки")
	}
	requireActiveVersion(t, &m, old.Tag)
	entries, err := os.ReadDir(m.root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("ожидались только прежняя версия, манифест и блокировка: %v", entries)
	}
}

func failedInstall(failure string, archive []byte, cancel context.CancelFunc) (http.RoundTripper, func(context.Context, Installation) error) {
	probe := func(context.Context, Installation) error {
		switch failure {
		case "probe":
			return errors.New("PoB не запустился")
		case "cancel_before_activation":
			cancel()
		}
		return nil
	}
	if failure == "download" {
		return archiveTransport(archive[:len(archive)/2]), probe
	}
	return archiveTransport(archive), probe
}
