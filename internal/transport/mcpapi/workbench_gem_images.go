package mcpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/poblegacy"
	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
)

const maxGemImage = 128 << 10

type gemImageReference struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type gemImages struct {
	references map[string]gemImageReference
	cache      string
	client     *http.Client
}

func newGemImages() (*gemImages, error) {
	content, err := poblegacy.Read("data/gem-icons.json")
	if err != nil {
		return nil, err
	}
	var manifest struct {
		Icons map[string]gemImageReference `json:"icons"`
	}
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, err
	}
	cache, _ := os.UserCacheDir()
	if cache != "" {
		cache = filepath.Join(cache, "pob2-engine", "gem-icons")
	}
	client := &http.Client{Timeout: 6 * time.Second, CheckRedirect: gemImageRedirect}
	return &gemImages{references: manifest.Icons, cache: cache, client: client}, nil
}

func gemImageRedirect(request *http.Request, previous []*http.Request) error {
	if !validGemImageURL(request.URL) || len(previous) >= 5 {
		return fmt.Errorf("недопустимое перенаправление для иконки камня")
	}
	return nil
}

func validGemImageURL(address *url.URL) bool {
	return address.Scheme == "https" && address.Host == "cdn.poe2db.tw" && address.User == nil &&
		strings.HasPrefix(address.Path, "/image/Art/") && address.RawQuery == "" && address.Fragment == ""
}

func (g *gemImages) read(ctx context.Context, id string) ([]byte, string, error) {
	entry := g.references[id]
	address, err := url.Parse(entry.URL)
	if err != nil || !validGemImageURL(address) || !validGemDigest(entry.SHA256) {
		return nil, "", fmt.Errorf("оригинальное изображение камня не найдено")
	}
	if data := g.cached(entry.SHA256); data != nil {
		return data, entry.URL, nil
	}
	data, err := g.download(ctx, entry)
	if err != nil {
		return nil, "", err
	}
	g.save(entry.SHA256, data)
	return data, entry.URL, nil
}

func validGemDigest(digest string) bool {
	decoded, err := hex.DecodeString(digest)
	return err == nil && len(decoded) == sha256.Size
}

func (g *gemImages) cached(digest string) []byte {
	if g.cache == "" {
		return nil
	}
	file, err := os.Open(filepath.Join(g.cache, digest+".webp"))
	if err != nil {
		return nil
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxGemImage+1))
	if err != nil || !validGemImage(data, digest) {
		return nil
	}
	return data
}

func (g *gemImages) download(ctx context.Context, entry gemImageReference) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, entry.URL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "pob2_engine")
	request.Header.Set("Referer", "https://poe2db.tw/")
	response, err := g.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("иконка камня недоступна: HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxGemImage+1))
	if err != nil {
		return nil, err
	}
	if !validGemImage(data, entry.SHA256) {
		return nil, fmt.Errorf("иконка камня не прошла проверку SHA-256 или формата WebP")
	}
	return data, nil
}

func validGemImage(data []byte, digest string) bool {
	return len(data) >= 12 && len(data) <= maxGemImage && bytes.Equal(data[:4], []byte("RIFF")) &&
		bytes.Equal(data[8:12], []byte("WEBP")) && fmt.Sprintf("%x", sha256.Sum256(data)) == digest
}

func (g *gemImages) save(digest string, data []byte) {
	if g.cache == "" || os.MkdirAll(g.cache, 0700) != nil {
		return
	}
	file, err := os.CreateTemp(g.cache, ".gem-*.webp")
	if err != nil {
		return
	}
	defer os.Remove(file.Name())
	_, err = file.Write(data)
	if closeErr := file.Close(); err == nil && closeErr == nil {
		_ = privatefs.Replace(file.Name(), filepath.Join(g.cache, digest+".webp"))
	}
}
