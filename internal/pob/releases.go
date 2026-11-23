package pob

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type release struct {
	Tag        string  `json:"tag_name"`
	Draft      bool    `json:"draft"`
	Prerelease bool    `json:"prerelease"`
	Assets     []asset `json:"assets"`
	commit     string
}

type asset struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Digest string `json:"digest"`
	url    string
}

const portableName = "PathOfBuildingCommunity-PoE2-Portable.zip"

// initialRelease соответствует эталону в testdata/reference/sources.lock.json.
func initialRelease() release {
	return release{
		Tag:    "v0.23.1",
		commit: "7d6f530cbdab20389ff8bc6ba97a37ac27f74e41",
		Assets: []asset{{
			Name:   portableName,
			Size:   382640331,
			Digest: "sha256:7e190d3b66191b09416a9cd4d12394d37df374d6d6e59ab4f2366f3d9e81baf7",
		}},
	}
}

func checkRedirect(req *http.Request, previous []*http.Request) error {
	req.Header.Del("Authorization")
	if len(previous) >= 6 || req.URL.Scheme != "https" || req.URL.User != nil || req.URL.Port() != "" {
		return fmt.Errorf("недопустимое перенаправление GitHub")
	}
	switch req.URL.Hostname() {
	case "github.com", "api.github.com", "objects.githubusercontent.com", "release-assets.githubusercontent.com":
		return nil
	default:
		return fmt.Errorf("недопустимый источник пакета PoB: %s", req.URL.Hostname())
	}
}

func (m *Manager) latest(ctx context.Context) (release, error) {
	return m.stableRelease(ctx, latestURL)
}

func (m *Manager) tagged(ctx context.Context, tag string) (release, error) {
	if !tags.MatchString(tag) {
		return release{}, fmt.Errorf("версия PoB должна иметь вид vX.Y.Z")
	}
	value, err := m.stableRelease(ctx, "https://api.github.com/repos/"+repository+"/releases/tags/"+tag)
	if err != nil {
		return release{}, err
	}
	if value.Tag != tag {
		return release{}, fmt.Errorf("GitHub вернул %s вместо %s", value.Tag, tag)
	}
	return value, nil
}

func (m *Manager) stableRelease(ctx context.Context, url string) (release, error) {
	var rel release
	if err := m.readJSON(ctx, url, &rel); err != nil {
		return rel, fmt.Errorf("проверка выпуска PoB: %w", err)
	}
	if !tags.MatchString(rel.Tag) || rel.Draft || rel.Prerelease {
		return rel, fmt.Errorf("GitHub вернул неподходящий тег PoB")
	}
	return rel, nil
}

func (m *Manager) readJSON(ctx context.Context, url string, result any) error {
	body, err := m.fetch(ctx, url)
	if err != nil {
		return err
	}
	defer body.Close()
	return json.NewDecoder(io.LimitReader(body, 1<<20)).Decode(result)
}

func (m *Manager) fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "pob2_engine")
	if m.githubToken != "" && strings.HasPrefix(url, "https://api.github.com/repos/"+repository+"/") {
		req.Header.Set("Authorization", "Bearer "+m.githubToken)
	}
	response, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("GitHub вернул HTTP %d", response.StatusCode)
	}
	return response.Body, nil
}

func portableAsset(rel release) (asset, error) {
	for _, file := range rel.Assets {
		if file.Name != portableName {
			continue
		}
		if !digestPattern.MatchString(file.sha256()) || file.Size <= 0 || file.Size > 1<<30 {
			return asset{}, fmt.Errorf("для портативного PoB нужны корректные размер и SHA-256")
		}
		file.url = downloadURL(repository, rel.Tag, portableName)
		return file, nil
	}
	return asset{}, fmt.Errorf("в выпуске %s отсутствует портативный архив PoB", rel.Tag)
}

func downloadURL(repository, tag, name string) string {
	return "https://github.com/" + repository + "/releases/download/" + tag + "/" + name
}

func (a asset) sha256() string {
	digest, _ := strings.CutPrefix(a.Digest, "sha256:")
	return digest
}
