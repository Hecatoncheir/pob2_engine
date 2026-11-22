package maintenance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	repository   = "Hecatoncheir/pob2_engine"
	latestURL    = "https://api.github.com/repos/" + repository + "/releases/latest"
	downloadBase = "https://github.com/" + repository + "/releases/download/"
	releaseURL   = "https://github.com/" + repository + "/releases/tag/"
)

type release struct {
	Tag        string  `json:"tag_name"`
	Draft      bool    `json:"draft"`
	Prerelease bool    `json:"prerelease"`
	Assets     []asset `json:"assets"`
}

type asset struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Digest string `json:"digest"`
}

func githubClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second, CheckRedirect: githubRedirect}
}

func githubRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 6 || req.URL.Scheme != "https" || req.URL.User != nil || req.URL.Port() != "" {
		return fmt.Errorf("недопустимое перенаправление GitHub")
	}
	switch req.URL.Hostname() {
	case "github.com", "api.github.com", "release-assets.githubusercontent.com", "objects.githubusercontent.com":
		return nil
	default:
		return fmt.Errorf("недопустимый узел загрузки GitHub: %s", req.URL.Hostname())
	}
}

func (s *Service) fetch(ctx context.Context, endpoint string, maxBytes int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "pob2_engine/"+s.version)
	response, err := s.client.Do(req)
	if err != nil {
		var requestError *url.Error
		if errors.As(err, &requestError) {
			return nil, requestError.Err
		}
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub вернул HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("превышен допустимый размер файла GitHub")
	}
	return data, err
}

func (s *Service) latest(ctx context.Context) (release, error) {
	data, err := s.fetch(ctx, latestURL, 1<<20)
	if err != nil {
		return release{}, fmt.Errorf("проверка стабильного выпуска: %w", err)
	}
	var value release
	if err := json.Unmarshal(data, &value); err != nil {
		return release{}, fmt.Errorf("чтение метаданных выпуска: %w", err)
	}
	if _, ok := parseVersion(value.Tag); !ok || value.Draft || value.Prerelease || strings.Contains(value.Tag, "-") {
		return release{}, fmt.Errorf("GitHub вернул неподходящий тег выпуска")
	}
	return value, nil
}

var versionPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*))?$`)

type version struct {
	numbers    [3]uint64
	prerelease bool
}

func parseVersion(tag string) (version, bool) {
	match := versionPattern.FindStringSubmatch(tag)
	if match == nil {
		return version{}, false
	}
	v := version{prerelease: match[4] != ""}
	for i := range v.numbers {
		number, err := strconv.ParseUint(match[i+1], 10, 64)
		if err != nil {
			return version{}, false
		}
		v.numbers[i] = number
	}
	return v, true
}

func (v version) olderThan(stable version) bool {
	for i, number := range v.numbers {
		if number != stable.numbers[i] {
			return number < stable.numbers[i]
		}
	}
	return v.prerelease
}
