package maintenance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func serveTestRelease(t *testing.T, s *Service, tag string, manifestPayload, payload []byte) *int {
	t.Helper()
	name, err := assetName(tag, s.goos, s.goarch)
	if err != nil {
		t.Fatal(err)
	}
	sums := []byte(sha256Hex(manifestPayload) + "  " + name + "\n")
	releaseData, err := json.Marshal(release{Tag: tag, Assets: []asset{
		{Name: name, Size: int64(len(manifestPayload)), Digest: "sha256:" + sha256Hex(manifestPayload)},
		{Name: "SHA256SUMS", Size: int64(len(sums)), Digest: "sha256:" + sha256Hex(sums)},
	}})
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{latestURL: releaseData, downloadBase + tag + "/SHA256SUMS": sums, downloadBase + tag + "/" + name: payload}
	requests := 0
	s.client = &http.Client{Transport: transportFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		data, ok := files[req.URL.String()]
		if !ok {
			return nil, fmt.Errorf("лишний запрос %s", req.URL)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(data)), Header: make(http.Header)}, nil
	})}
	return &requests
}

func testInstallation(t *testing.T) *Service {
	t.Helper()
	path := filepath.Join(t.TempDir(), "engine")
	if err := os.WriteFile(path, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, name := range []string{path + ".update.lock", path} {
			if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
				t.Error(err)
			}
		}
	})
	image, err := snapshotFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return &Service{version: "v0.9.0", goos: "linux", goarch: "amd64", path: path, image: image}
}

func requireInstalled(t *testing.T, s *Service, expected string) {
	t.Helper()
	data, err := os.ReadFile(s.path)
	if err != nil || string(data) != expected {
		t.Fatalf("установленный файл = %q, ошибка %v", data, err)
	}
}

func TestUpdateCheckAndInstall(t *testing.T) {
	s := testInstallation(t)
	requests := serveTestRelease(t, s, "v0.10.0", []byte("new"), []byte("new"))
	result, err := s.Update(t.Context(), false, "engine")
	if err != nil || result.Status != "available" || *requests != 1 {
		t.Fatalf("проверка = %+v, запросов %d, ошибка %v", result, *requests, err)
	}
	requireInstalled(t, s, "old")
	result, err = s.Update(t.Context(), true, "engine")
	if err != nil || result.Status != "updated" || !result.RestartRequired {
		t.Fatalf("обновление = %+v, ошибка %v", result, err)
	}
	requireInstalled(t, s, "new")
	report, err := s.Doctor(t.Context())
	if err != nil || !report.RestartRequired {
		t.Fatalf("диагностика = %+v, ошибка %v", report, err)
	}
	if _, err := s.Update(t.Context(), true, "engine"); err == nil {
		t.Fatal("процесс со старым образом повторно заменил бинарник")
	}
}

func TestUpdatePreservesBinaryOnInvalidPayload(t *testing.T) {
	s := testInstallation(t)
	serveTestRelease(t, s, "v0.10.0", []byte("new"), []byte("bad"))
	if _, err := s.Update(t.Context(), true, "engine"); err == nil {
		t.Fatal("принят повреждённый бинарник")
	}
	requireInstalled(t, s, "old")
}

func TestUpdatePreventsDowngradeAndUpdatingDevelopmentBuild(t *testing.T) {
	s := testInstallation(t)
	requests := serveTestRelease(t, s, "v0.8.0", []byte("new"), []byte("new"))
	result, err := s.Update(t.Context(), true, "engine")
	if err != nil || result.Status != "up_to_date" || *requests != 1 {
		t.Fatalf("проверка = %+v, запросов %d, ошибка %v", result, *requests, err)
	}
	s.version = "dev"
	if _, err := s.Update(t.Context(), true, "engine"); err == nil || *requests != 1 {
		t.Fatalf("обновление dev: запросов %d, ошибка %v", *requests, err)
	}
	requireInstalled(t, s, "old")
}

func TestUpdateKeepsBinaryWhenCanceled(t *testing.T) {
	s := testInstallation(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := s.install(ctx, []byte("new")); err != context.Canceled {
		t.Fatalf("ожидалась отмена, получено %v", err)
	}
	requireInstalled(t, s, "old")
}

func TestUpdateHonorsInstallationLock(t *testing.T) {
	s := testInstallation(t)
	requests := serveTestRelease(t, s, "v0.10.0", []byte("new"), []byte("new"))
	lock, err := privatefs.TryLock(s.path + ".update.lock")
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	if _, err := s.Update(t.Context(), true, "engine"); !errors.Is(err, privatefs.ErrLocked) || *requests != 1 {
		t.Fatalf("конкурентное обновление: запросов %d, ошибка %v", *requests, err)
	}
	requireInstalled(t, s, "old")
}
