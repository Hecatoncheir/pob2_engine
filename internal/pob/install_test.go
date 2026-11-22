package pob

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type installTransport func(*http.Request) (*http.Response, error)

func (f installTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestInitialPoBMatchesReference(t *testing.T) {
	data, err := os.ReadFile("../../testdata/reference/sources.lock.json")
	if err != nil {
		t.Fatal(err)
	}
	var lock struct {
		PoB struct {
			Tag    string `json:"tag"`
			Commit string `json:"commit"`
		} `json:"pob"`
	}
	if err := json.Unmarshal(data, &lock); err != nil {
		t.Fatal(err)
	}
	release := initialRelease()
	if release.Tag != lock.PoB.Tag || release.commit != lock.PoB.Commit {
		t.Fatalf("первая установка PoB %s (%s) не совпадает с эталоном", release.Tag, release.commit)
	}
	if _, err := portableAsset(release); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureDownloadsPinnedPoB(t *testing.T) {
	var urls []string
	m := Manager{root: t.TempDir(), client: &http.Client{Transport: installTransport(func(req *http.Request) (*http.Response, error) {
		urls = append(urls, req.URL.String())
		return archiveTransport("неверный архив").RoundTrip(req)
	})}}
	if _, err := m.Ensure(t.Context()); err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("повреждённый пакет должен быть отклонён: %v", err)
	}
	expected := downloadURL(repository, initialRelease().Tag, portableName)
	if len(urls) != 1 || urls[0] != expected {
		t.Fatalf("ожидалась прямая загрузка %s, запрошены %v", expected, urls)
	}
	if _, err := os.Stat(filepath.Join(m.root, "current.json")); !os.IsNotExist(err) {
		t.Fatalf("после сбоя активной установки быть не должно: %v", err)
	}
}

func TestEnsureKeepsInstalledPoBOffline(t *testing.T) {
	var requests int
	m := Manager{root: t.TempDir(), client: &http.Client{Transport: installTransport(func(*http.Request) (*http.Response, error) {
		requests++
		return nil, fmt.Errorf("сеть недоступна")
	})}}
	dir := filepath.Join(m.root, "staged")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	cached := Installation{Version: "v0.23.2", SHA256: strings.Repeat("1", 64), Commit: strings.Repeat("2", 40)}
	if err := m.publish(t.Context(), dir, cached); err != nil {
		t.Fatal(err)
	}
	active, err := m.Ensure(t.Context())
	if err != nil || active.Version != cached.Version || active.Commit != cached.Commit || requests != 0 {
		t.Fatalf("повторный запуск должен использовать кэш: %+v, обращений %d, ошибка %v", active, requests, err)
	}
}

func TestCheckUpdateStillQueriesLatestPoB(t *testing.T) {
	const next = "v0.23.2"
	m := Manager{root: t.TempDir(), client: &http.Client{Transport: installTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != latestURL {
			t.Fatalf("ожидалась проверка последнего выпуска, запрошен %s", req.URL)
		}
		return archiveTransport(`{"tag_name":"` + next + `"}`).RoundTrip(req)
	})}}
	state, err := m.Update(t.Context(), false)
	if err != nil || state.Status != "available" || state.LatestVersion != next {
		t.Fatalf("должен быть доступен новый выпуск: %+v, %v", state, err)
	}
}
