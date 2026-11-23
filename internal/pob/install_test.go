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
	if err := m.publish(t.Context(), dir, installationState{Installation: cached}); err != nil {
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

func TestCheckUpdateToUsesExactOfficialTag(t *testing.T) {
	const tag = "v0.23.1"
	url := "https://api.github.com/repos/" + repository + "/releases/tags/" + tag
	m := Manager{root: t.TempDir(), client: &http.Client{Transport: installTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != url {
			t.Fatalf("ожидалась проверка заданного выпуска, получен %s", req.URL)
		}
		return archiveTransport(`{"tag_name":"` + tag + `"}`).RoundTrip(req)
	})}}
	state, err := m.UpdateTo(t.Context(), false, tag)
	if err != nil || state.LatestVersion != tag || state.Status != "available" {
		t.Fatalf("не найден указанный выпуск: %+v, %v", state, err)
	}
}

func TestPoBGitHubTokenOnlyForOfficialMetadata(t *testing.T) {
	const token = "test-pob-token"
	t.Setenv("POB2_ENGINE_POB_GITHUB_TOKEN", token)
	m, err := Local()
	if err != nil {
		t.Fatal(err)
	}
	archive := downloadURL(repository, "v0.23.0", portableName)
	redirect := "https://api.github.com/repos/another/project/releases/latest"
	vectors := map[string]string{latestURL: "Bearer " + token, archive: "", redirect: ""}
	m.client.Transport = installTransport(func(req *http.Request) (*http.Response, error) {
		address := req.URL.String()
		want, ok := vectors[address]
		if !ok || req.Header.Get("Authorization") != want {
			t.Fatalf("неверная авторизация запроса %s", address)
		}
		if address == latestURL {
			return &http.Response{StatusCode: http.StatusFound, Header: http.Header{"Location": {redirect}}}, nil
		}
		return archiveTransport("{}").RoundTrip(req)
	})
	for _, address := range []string{latestURL, archive, redirect} {
		body, err := m.fetch(t.Context(), address)
		if err != nil {
			t.Fatal(err)
		}
		_ = body.Close()
	}
}
