package maintenance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
)

const (
	UpdateCapability = "maintenance.update"
	UpdateTimeout    = 6 * time.Minute
)

type UpdateResult struct {
	Status          string     `json:"status"`
	CurrentVersion  string     `json:"current_version"`
	LatestVersion   string     `json:"latest_version"`
	ReleaseURL      string     `json:"release_url"`
	RestartRequired bool       `json:"restart_required"`
	Engine          *pob.State `json:"engine,omitempty"`
	PoB             *pob.State `json:"pob,omitempty"`
}

func (s *Service) Update(ctx context.Context, apply bool, targets ...string) (UpdateResult, error) {
	target := ""
	if len(targets) != 0 {
		target = targets[0]
	}
	if target == "pob" || target == "pob-rollback" || strings.HasPrefix(target, "pob@") {
		state, err := s.updatePoB(ctx, apply, target)
		return UpdateResult{Status: state.Status, RestartRequired: state.RestartRequired, PoB: &state}, err
	}
	switch target {
	case "engine":
		return s.updateEngine(ctx, apply)
	case "", "all":
		return s.updateAll(ctx, apply)
	default:
		return UpdateResult{}, fmt.Errorf("цель обновления: all, engine, pob, pob@vX.Y.Z или pob-rollback")
	}
}

func (s *Service) updateEngine(ctx context.Context, apply bool) (UpdateResult, error) {
	current, isRelease := parseVersion(s.version)
	if apply && !isRelease {
		return UpdateResult{}, fmt.Errorf("самообновление доступно для сборок с релизным тегом")
	}
	latest, err := s.latest(ctx)
	if err != nil {
		return UpdateResult{}, err
	}
	result := UpdateResult{CurrentVersion: s.version, LatestVersion: latest.Tag, ReleaseURL: releaseURL + latest.Tag, Status: "up_to_date"}
	stable, _ := parseVersion(latest.Tag)
	if !isRelease {
		result.Status = "development_build"
		return result, nil
	}
	if !current.olderThan(stable) {
		return result, nil
	}
	result.Status = "available"
	if !apply {
		return result, nil
	}
	if err := s.apply(ctx, latest); err != nil {
		return UpdateResult{}, err
	}
	result.Status, result.RestartRequired = "updated", true
	return result, nil
}

func assetName(tag, goos, goarch string) (string, error) {
	name := "pob2_engine_" + tag + "_" + goos + "_" + goarch
	switch goos + "/" + goarch {
	case "windows/amd64":
		return name + ".exe", nil
	case "linux/amd64", "darwin/arm64":
		return name, nil
	default:
		return "", fmt.Errorf("для %s/%s бинарник обновления не выпускается", goos, goarch)
	}
}

func (s *Service) apply(ctx context.Context, latest release) error {
	name, err := assetName(latest.Tag, s.goos, s.goarch)
	if err != nil {
		return err
	}
	if s.pathErr != nil {
		return s.pathErr
	}
	lock, err := privatefs.TryLock(s.path + ".update.lock")
	if err != nil {
		return fmt.Errorf("блокировка обновления: %w", err)
	}
	defer lock.Close()
	if err := s.checkImage(); err != nil {
		return err
	}
	data, err := s.verifiedBinary(ctx, latest, name)
	if err != nil {
		return err
	}
	return s.install(ctx, data)
}

func (s *Service) checkImage() error {
	image, err := os.Lstat(s.path)
	if err != nil {
		return fmt.Errorf("проверка установленного бинарника: %w", err)
	}
	if !image.Mode().IsRegular() || !os.SameFile(image, s.image) {
		return fmt.Errorf("установленный бинарник изменился. Перезапустите процесс перед обновлением")
	}
	return nil
}

func (s *Service) verifiedBinary(ctx context.Context, latest release, name string) ([]byte, error) {
	sums, err := s.downloadAsset(ctx, latest, "SHA256SUMS", 64<<10)
	if err != nil {
		return nil, err
	}
	expected, err := checksum(sums, name)
	if err != nil {
		return nil, err
	}
	binary, err := s.downloadAsset(ctx, latest, name, 64<<20)
	if err != nil {
		return nil, err
	}
	if sha256Hex(binary) != expected {
		return nil, fmt.Errorf("SHA-256 %s не совпадает с SHA256SUMS", name)
	}
	return binary, nil
}

func checksum(data []byte, name string) (string, error) {
	var expected string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != name {
			continue
		}
		if expected != "" || !validDigest(fields[0]) {
			return "", fmt.Errorf("некорректная запись SHA256SUMS для %s", name)
		}
		expected = fields[0]
	}
	if expected == "" {
		return "", fmt.Errorf("в SHA256SUMS нет записи для %s", name)
	}
	return expected, nil
}

func (s *Service) downloadAsset(ctx context.Context, latest release, name string, limit int64) ([]byte, error) {
	a, err := releaseAsset(latest, name, limit)
	if err != nil {
		return nil, err
	}
	data, err := s.fetch(ctx, downloadBase+latest.Tag+"/"+name, a.Size)
	if err != nil {
		return nil, fmt.Errorf("скачивание %s: %w", name, err)
	}
	if int64(len(data)) != a.Size || "sha256:"+sha256Hex(data) != a.Digest {
		return nil, fmt.Errorf("размер или SHA-256 %s не совпадает с метаданными GitHub", name)
	}
	return data, nil
}

func releaseAsset(latest release, name string, limit int64) (asset, error) {
	var found asset
	for _, a := range latest.Assets {
		if a.Name != name {
			continue
		}
		digest, ok := strings.CutPrefix(a.Digest, "sha256:")
		if found.Name != "" || a.Size <= 0 || a.Size > limit || !ok || !validDigest(digest) {
			return asset{}, fmt.Errorf("некорректные метаданные файла %s", name)
		}
		found = a
	}
	if found.Name == "" {
		return asset{}, fmt.Errorf("в выпуске %s нет файла %s", latest.Tag, name)
	}
	return found, nil
}

func validDigest(digest string) bool {
	raw, err := hex.DecodeString(digest)
	return err == nil && len(raw) == sha256.Size
}

func sha256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (s *Service) install(ctx context.Context, data []byte) error {
	temp, err := os.CreateTemp(filepath.Dir(s.path), ".pob2-engine-update-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err := writeExecutable(temp, data, s.image.Mode().Perm()); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.checkImage(); err != nil {
		return err
	}
	if err := replaceExecutable(temp.Name(), s.path); err != nil {
		return fmt.Errorf("замена установленного бинарника: %w", err)
	}
	return nil
}

func writeExecutable(file *os.File, data []byte, permissions os.FileMode) error {
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Chmod(permissions); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	return file.Close()
}
