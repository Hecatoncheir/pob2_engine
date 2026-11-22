package pob

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

type archiveTransport []byte

func (a archiveTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(a))}, nil
}

func tinyArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var data bytes.Buffer
	w := zip.NewWriter(&data)
	for name, content := range files {
		entry, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(entry, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func TestExtractVerifiedArchive(t *testing.T) {
	contents := tinyArchive(t, map[string]string{"src/Launch.lua": "return true", "PoB.exe": "GUI"})
	m := Manager{root: t.TempDir(), client: &http.Client{Transport: archiveTransport(contents)}}
	a := asset{Name: "pob.zip", url: "https://example.invalid/pob.zip", Size: int64(len(contents)),
		Digest: fmt.Sprintf("sha256:%x", sha256.Sum256(contents))}
	dir := t.TempDir()
	if err := m.extract(t.Context(), a, dir, false); err != nil {
		t.Fatal(err)
	}
	if !regularFile(filepath.Join(dir, "src", "Launch.lua")) || regularFile(filepath.Join(dir, "PoB.exe")) {
		t.Fatal("неверный состав установленного PoB")
	}
	a.Digest = "sha256:" + fmt.Sprintf("%064d", 0)
	if err := m.extract(t.Context(), a, t.TempDir(), false); err == nil {
		t.Fatal("принят архив с неверной контрольной суммой")
	}
}

func TestRejectPoBArchiveTraversal(t *testing.T) {
	contents := tinyArchive(t, map[string]string{"../escape.lua": "return true"})
	archive, err := zip.NewReader(bytes.NewReader(contents), int64(len(contents)))
	if err != nil {
		t.Fatal(err)
	}
	if err := extractFiles(context.Background(), archive.File, t.TempDir(), false); err == nil {
		t.Fatal("принят путь за пределами архива")
	}
}

func TestPublishInstallation(t *testing.T) {
	m := Manager{root: t.TempDir()}
	staging, err := os.MkdirTemp(m.root, "stage-")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staging, "Launch.lua"), []byte("return true"), 0600); err != nil {
		t.Fatal(err)
	}
	i := Installation{Version: "v0.23.1", SHA256: fmt.Sprintf("%064d", 1), Commit: fmt.Sprintf("%040d", 1)}
	if err := m.publish(t.Context(), staging, installationState{Installation: i}); err != nil {
		t.Fatal(err)
	}
	active, err := m.Installed()
	if err != nil || active.Version != i.Version || !regularFile(filepath.Join(active.Directory, "Launch.lua")) {
		t.Fatalf("неверная активная установка: %+v, %v", active, err)
	}
}
