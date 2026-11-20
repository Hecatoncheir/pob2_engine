// reference проверяет закреплённые источники будущих эталонов без запуска Lua.
package main

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

const maxFileSize = 2 << 20

type fingerprint struct {
	Path    string `json:"path"`
	GitBlob string `json:"git_blob"`
	SHA256  string `json:"sha256"`
}

type source struct {
	Repository string        `json:"repository"`
	Tag        string        `json:"tag,omitempty"`
	Commit     string        `json:"commit"`
	Files      []fingerprint `json:"files"`
}

type manifest struct {
	Schema               int    `json:"schema"`
	CheckedAt            string `json:"checked_at"`
	TreeVersion          string `json:"tree_version"`
	MCPEmbeddedPoBCommit string `json:"mcp_embedded_pob_commit"`
	PoB                  source `json:"pob"`
	MCP                  source `json:"mcp"`
	PairEvidence         source `json:"pair_evidence"`
}

var repositoryName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*/[A-Za-z0-9][A-Za-z0-9_.-]*$`)
var commitHash = regexp.MustCompile(`^[a-f0-9]{40}$`)
var checksum = regexp.MustCompile(`^[a-f0-9]{64}$`)
var versionTag = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

func main() {
	lockPath := flag.String("lock", "testdata/reference/sources.lock.json", "манифест закреплённых источников")
	remote := flag.Bool("remote", false, "сверить исходные файлы по HTTPS и проверить происхождение пары")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "неожиданные позиционные аргументы")
		os.Exit(2)
	}
	if err := run(*lockPath, *remote); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(lockPath string, remote bool) error {
	file, err := os.Open(lockPath)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxFileSize+1))
	if err != nil {
		return err
	}
	m, err := load(data)
	if err != nil {
		return err
	}
	if remote {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if err := verifyRemote(ctx, &http.Client{Timeout: 15 * time.Second}, m); err != nil {
			return err
		}
		fmt.Println("Контрольные суммы и происхождение пары подтверждены. Эталонный расчёт Lua ещё проверяется отдельно.")
	}
	fmt.Printf("PoB %s (%s); MCP %s (%s); дерево %s.\n", m.PoB.Tag, m.PoB.Commit, m.MCP.Tag, m.MCP.Commit, m.TreeVersion)
	return nil
}

func load(data []byte) (manifest, error) {
	var m manifest
	if len(data) > maxFileSize {
		return m, fmt.Errorf("манифест слишком велик")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&m); err != nil {
		return m, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return m, fmt.Errorf("лишние данные после манифеста")
	}
	if m.Schema != 1 || m.TreeVersion == "" || !commitHash.MatchString(m.MCPEmbeddedPoBCommit) {
		return m, fmt.Errorf("неверная схема или версия источников")
	}
	if _, err := time.Parse("2006-01-02", m.CheckedAt); err != nil {
		return m, err
	}
	if !versionTag.MatchString(m.PoB.Tag) || !versionTag.MatchString(m.MCP.Tag) {
		return m, fmt.Errorf("нужны явные версии PoB и MCP")
	}
	for _, item := range []source{m.PoB, m.MCP, m.PairEvidence} {
		if err := validateSource(item); err != nil {
			return m, err
		}
	}
	return m, nil
}

func validateSource(s source) error {
	if !repositoryName.MatchString(s.Repository) || !commitHash.MatchString(s.Commit) || len(s.Files) == 0 {
		return fmt.Errorf("неполный закреплённый источник %q", s.Repository)
	}
	seen := make(map[string]bool)
	for _, f := range s.Files {
		if f.Path == "." || path.IsAbs(f.Path) || path.Clean(f.Path) != f.Path || strings.HasPrefix(f.Path, "../") || strings.ContainsAny(f.Path, `\?#%`) {
			return fmt.Errorf("некорректный путь источника %q", f.Path)
		}
		if seen[f.Path] || !commitHash.MatchString(f.GitBlob) || !checksum.MatchString(f.SHA256) {
			return fmt.Errorf("повтор пути или неверная контрольная сумма %s", f.Path)
		}
		seen[f.Path] = true
	}
	return nil
}

func fetch(ctx context.Context, client *http.Client, s source, f fingerprint) ([]byte, error) {
	address := "https://raw.githubusercontent.com/" + s.Repository + "/" + s.Commit + "/" + f.Path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d", address, response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxFileSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxFileSize {
		return nil, fmt.Errorf("файл источника слишком велик: %s", f.Path)
	}
	// SHA-1 нужен для сопоставления с Git; целостность дополнительно закреплена SHA-256.
	blob := sha1.New()
	fmt.Fprintf(blob, "blob %d\x00", len(data))
	blob.Write(data)
	if fmt.Sprintf("%x", sha256.Sum256(data)) != f.SHA256 || fmt.Sprintf("%x", blob.Sum(nil)) != f.GitBlob {
		return nil, fmt.Errorf("контрольная сумма не совпала: %s/%s", s.Repository, f.Path)
	}
	return data, nil
}

func verifyRemote(ctx context.Context, client *http.Client, m manifest) error {
	contents := make(map[string][]byte)
	for _, s := range []source{m.PoB, m.MCP, m.PairEvidence} {
		for _, f := range s.Files {
			data, err := fetch(ctx, client, s, f)
			if err != nil {
				return err
			}
			contents[s.Repository+"/"+f.Path] = data
		}
	}
	var pair struct{ PoB, MCP struct{ Tag, Commit string } }
	if err := json.Unmarshal(contents[m.PairEvidence.Repository+"/tool/components-lock.json"], &pair); err != nil {
		return err
	}
	if pair.PoB.Commit != m.PoB.Commit || pair.MCP.Commit != m.MCP.Commit || pair.PoB.Tag != m.PoB.Tag || pair.MCP.Tag != m.MCP.Tag {
		return fmt.Errorf("зафиксированная пара не совпадает с источником её происхождения")
	}
	var embedded struct{ Repository, Revision, TreeVersion string }
	if err := json.Unmarshal(contents[m.MCP.Repository+"/pob_poe2_mcp/data/engine-lock.json"], &embedded); err != nil {
		return err
	}
	if embedded.Revision != m.MCPEmbeddedPoBCommit || embedded.TreeVersion != m.TreeVersion || embedded.Repository != "https://github.com/"+m.PoB.Repository+".git" {
		return fmt.Errorf("встроенный pin MCP изменился")
	}
	return nil
}
