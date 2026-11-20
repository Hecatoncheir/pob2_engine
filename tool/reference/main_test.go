package main

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fixture() (manifest, map[string][]byte) {
	m := manifest{Schema: 1, CheckedAt: "2026-11-20", TreeVersion: "0_5", MCPEmbeddedPoBCommit: strings.Repeat("d", 40),
		PoB:          source{Repository: "example/pob", Tag: "v0.23.1", Commit: strings.Repeat("a", 40)},
		MCP:          source{Repository: "example/mcp", Tag: "v0.24.1", Commit: strings.Repeat("b", 40)},
		PairEvidence: source{Repository: "example/ui", Commit: strings.Repeat("c", 40)}}
	bodies := map[string][]byte{
		"example/pob/src/GameVersions.lua":               []byte(`treeVersionList = { "0_5" }`),
		"example/mcp/pob_poe2_mcp/data/engine-lock.json": []byte(fmt.Sprintf(`{"repository":"https://github.com/example/pob.git","revision":%q,"treeVersion":"0_5"}`, m.MCPEmbeddedPoBCommit)),
		"example/ui/tool/components-lock.json":           []byte(fmt.Sprintf(`{"pob":{"tag":%q,"commit":%q},"mcp":{"tag":%q,"commit":%q}}`, m.PoB.Tag, m.PoB.Commit, m.MCP.Tag, m.MCP.Commit)),
	}
	m.PoB.Files = []fingerprint{fingerprintOf("src/GameVersions.lua", bodies["example/pob/src/GameVersions.lua"])}
	m.MCP.Files = []fingerprint{fingerprintOf("pob_poe2_mcp/data/engine-lock.json", bodies["example/mcp/pob_poe2_mcp/data/engine-lock.json"])}
	m.PairEvidence.Files = []fingerprint{fingerprintOf("tool/components-lock.json", bodies["example/ui/tool/components-lock.json"])}
	return m, bodies
}

func fingerprintOf(name string, data []byte) fingerprint {
	blob := append([]byte(fmt.Sprintf("blob %d\x00", len(data))), data...)
	return fingerprint{Path: name, GitBlob: fmt.Sprintf("%x", sha1.Sum(blob)), SHA256: fmt.Sprintf("%x", sha256.Sum256(data))}
}

func fakeClient(bodies map[string][]byte) *http.Client {
	return &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Scheme != "https" || r.URL.Host != "raw.githubusercontent.com" {
			return nil, fmt.Errorf("неожиданный адрес")
		}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(parts) < 4 {
			return nil, fmt.Errorf("неполный адрес источника")
		}
		pins := map[string]string{"example/pob": strings.Repeat("a", 40), "example/mcp": strings.Repeat("b", 40), "example/ui": strings.Repeat("c", 40)}
		if parts[2] != pins[strings.Join(parts[:2], "/")] {
			return nil, fmt.Errorf("запрос использует незакреплённую ревизию")
		}
		key := strings.Join(append(parts[:2], parts[3:]...), "/")
		data, ok := bodies[key]
		if !ok {
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(""))}, nil
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(data)))}, nil
	})}
}

func TestManifestRejectsFloatingReferencesAndUnsafePaths(t *testing.T) {
	for _, mutate := range []func(*manifest){
		func(m *manifest) { m.PoB.Commit = "main" },
		func(m *manifest) { m.MCP.Files[0].Path = "../private" },
		func(m *manifest) { m.MCP.Files[0].SHA256 = "missing" },
		func(m *manifest) { m.PairEvidence.Repository = "https://example.com/repo" },
		func(m *manifest) { m.PoB.Files = append(m.PoB.Files, m.PoB.Files[0]) },
		func(m *manifest) { m.Schema = 99 },
	} {
		m, _ := fixture()
		mutate(&m)
		data, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := load(data); err == nil {
			t.Fatal("некорректный манифест принят")
		}
	}
	m, _ := fixture()
	data, _ := json.Marshal(m)
	if _, err := load(data); err != nil {
		t.Fatal(err)
	}
	if _, err := load(append(data, []byte(` {"schema":1}`)...)); err == nil {
		t.Fatal("лишний JSON принят")
	}
}

func TestRemoteChecksContentsAndBothDigests(t *testing.T) {
	m, bodies := fixture()
	if err := verifyRemote(t.Context(), fakeClient(bodies), m); err != nil {
		t.Fatal(err)
	}
	for _, change := range []string{"body", "git", "sha256"} {
		m, bodies := fixture()
		switch change {
		case "body":
			bodies["example/pob/src/GameVersions.lua"] = []byte("changed")
		case "git":
			m.PoB.Files[0].GitBlob = strings.Repeat("0", 40)
		case "sha256":
			m.PoB.Files[0].SHA256 = strings.Repeat("0", 64)
		}
		if err := verifyRemote(t.Context(), fakeClient(bodies), m); err == nil {
			t.Fatalf("изменение %s не обнаружено", change)
		}
	}
}

func TestRemoteRejectsInconsistentPairEvenWithMatchingDigests(t *testing.T) {
	m, bodies := fixture()
	key := "example/ui/tool/components-lock.json"
	bodies[key] = []byte(strings.ReplaceAll(string(bodies[key]), m.PoB.Commit, strings.Repeat("e", 40)))
	m.PairEvidence.Files[0] = fingerprintOf("tool/components-lock.json", bodies[key])
	if err := verifyRemote(t.Context(), fakeClient(bodies), m); err == nil {
		t.Fatal("источник подтверждает другую пару")
	}
	m, bodies = fixture()
	m.MCPEmbeddedPoBCommit = strings.Repeat("e", 40)
	if err := verifyRemote(t.Context(), fakeClient(bodies), m); err == nil {
		t.Fatal("различие встроенного pin не обнаружено")
	}
}

func TestFetchRejectsMissingOversizedAndCanceledResponses(t *testing.T) {
	m, _ := fixture()
	if _, err := fetch(t.Context(), fakeClient(nil), m.PoB, m.PoB.Files[0]); err == nil {
		t.Fatal("HTTP 404 принят")
	}
	client := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", maxFileSize+1)))}, nil
	})}
	if _, err := fetch(t.Context(), client, m.PoB, m.PoB.Files[0]); err == nil || !strings.Contains(err.Error(), "велик") {
		t.Fatalf("предел размера: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	client.Transport = roundTrip(func(r *http.Request) (*http.Response, error) { return nil, r.Context().Err() })
	if _, err := fetch(ctx, client, m.PoB, m.PoB.Files[0]); !errors.Is(err, context.Canceled) {
		t.Fatalf("отмена: %v", err)
	}
}
