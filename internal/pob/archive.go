package pob

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
)

func (m *Manager) install(ctx context.Context, rel release) error {
	file, err := portableAsset(rel)
	if err != nil {
		return err
	}
	revision, err := m.releaseCommit(ctx, rel)
	if err != nil {
		return err
	}
	return m.stage(ctx, file, Installation{Version: rel.Tag, SHA256: file.sha256(), Commit: revision})
}

func (m *Manager) releaseCommit(ctx context.Context, rel release) (string, error) {
	if rel.commit != "" {
		return rel.commit, nil
	}
	var revision struct {
		SHA string `json:"sha"`
	}
	if err := m.readJSON(ctx, "https://api.github.com/repos/"+repository+"/commits/"+rel.Tag, &revision); err != nil {
		return "", err
	}
	if !commitPattern.MatchString(revision.SHA) {
		return "", fmt.Errorf("не удалось определить commit выпуска PoB")
	}
	return revision.SHA, nil
}

func (m *Manager) stage(ctx context.Context, file asset, i Installation) error {
	staging, err := os.MkdirTemp(m.root, ".pob-stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	if err := m.extract(ctx, file, staging, false); err != nil {
		return err
	}
	if m.nativeRuntime.url != "" {
		if err := m.extract(ctx, m.nativeRuntime, staging, true); err != nil {
			return err
		}
		i.NativeDirectory = filepath.Join(staging, "native")
	}
	i.Directory, err = sourceRoot(staging)
	if err != nil {
		return err
	}
	if err := m.probe(ctx, i); err != nil {
		return fmt.Errorf("новый PoB не прошёл проверку запуска: %w", err)
	}
	return m.publish(ctx, staging, i)
}

func sourceRoot(staging string) (string, error) {
	for _, candidate := range []string{staging, filepath.Join(staging, "src")} {
		if regularFile(filepath.Join(candidate, "Launch.lua")) {
			return staging, nil
		}
	}
	return "", fmt.Errorf("в портативном PoB отсутствует Launch.lua")
}

func (m *Manager) publish(ctx context.Context, staging string, i Installation) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := moveInstallation(staging, m.locate(i)); err != nil {
		return err
	}
	f, err := os.CreateTemp(m.root, ".pob-current-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err := writeManifest(f, i); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return privatefs.Replace(f.Name(), filepath.Join(m.root, "current.json"))
}

func moveInstallation(staging string, target Installation) error {
	_, err := os.Stat(target.Directory)
	switch {
	case os.IsNotExist(err):
		return os.Rename(staging, target.Directory)
	case err != nil:
		return err
	default:
		_, err := target.Layout()
		return err
	}
}

func writeManifest(f *os.File, i Installation) error {
	defer f.Close()
	if err := json.NewEncoder(f).Encode(i); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	return f.Close()
}

func (m *Manager) extract(ctx context.Context, a asset, dir string, native bool) error {
	file, err := os.CreateTemp(m.root, ".pob-archive-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err := m.download(ctx, a, file); err != nil {
		return err
	}
	archive, err := zip.NewReader(file, a.Size)
	if err != nil {
		return err
	}
	return extractFiles(ctx, archive.File, dir, native)
}

func (m *Manager) download(ctx context.Context, a asset, file *os.File) error {
	body, err := m.fetch(ctx, a.url)
	if err != nil {
		return err
	}
	defer body.Close()
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(file, hash), io.LimitReader(body, a.Size+1))
	if err != nil {
		return err
	}
	if size != a.Size || hex.EncodeToString(hash.Sum(nil)) != a.sha256() {
		return fmt.Errorf("размер или SHA-256 %s не совпадает с опубликованным", a.Name)
	}
	return nil
}

func extractFiles(ctx context.Context, files []*zip.File, directory string, native bool) error {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	var total uint64
	for _, entry := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		size, err := extractEntry(root, entry, native, total)
		if err != nil {
			return err
		}
		total += size
	}
	return nil
}

func extractEntry(root *os.Root, entry *zip.File, native bool, total uint64) (uint64, error) {
	name, err := safeArchivePath(entry.Name)
	if err != nil {
		return 0, err
	}
	if entry.FileInfo().IsDir() || !needsFile(name, native) {
		return 0, nil
	}
	size := entry.UncompressedSize64
	if size > 128<<20 || total+size > 3<<30 {
		return 0, fmt.Errorf("архив PoB превышает лимит распаковки")
	}
	return size, extractFile(root, entry, name)
}

func safeArchivePath(name string) (string, error) {
	name = strings.TrimSuffix(name, "/")
	local := filepath.FromSlash(name)
	if !filepath.IsLocal(local) || name != path.Clean(name) || strings.ContainsAny(name, "\\:") {
		return "", fmt.Errorf("недопустимый путь в архиве PoB: %q", name)
	}
	return local, nil
}

func needsFile(name string, native bool) bool {
	name = filepath.ToSlash(name)
	if native {
		return strings.HasPrefix(name, "native/")
	}
	first, _, _ := strings.Cut(name, "/")
	switch first {
	case "src", "runtime", "Assets", "Classes", "Data", "Export", "Modules", "TreeData", "lua":
		return !strings.EqualFold(path.Ext(name), ".exe")
	}
	ext := strings.ToLower(path.Ext(name))
	license := strings.HasPrefix(strings.ToLower(name), "license")
	return ext == ".lua" || ext == ".dll" || license || name == "manifest.xml"
}

func extractFile(root *os.Root, entry *zip.File, name string) error {
	if !entry.Mode().IsRegular() {
		return fmt.Errorf("архив PoB содержит ссылку или специальный файл: %s", name)
	}
	if err := root.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return err
	}
	in, err := entry.Open()
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	written, err := io.Copy(out, io.LimitReader(in, int64(entry.UncompressedSize64)+1))
	closeErr := out.Close()
	if err != nil {
		return err
	}
	if written != int64(entry.UncompressedSize64) {
		return fmt.Errorf("неверный размер файла %s", name)
	}
	return closeErr
}

func nativeAsset(goos, arch string) asset {
	const repo, tag = "Hecatoncheir/pob-poe2-mcp", "v0.24.1"
	var file asset
	switch goos + "/" + arch {
	case "linux/amd64":
		file = asset{Name: "pob-poe2-mcp-native-linux-x64.zip", Size: 817929, Digest: "sha256:e7dbdb4319c417df4ea5365b3a3d254aacf8cec8deabb7a0f809f61e08a5c1a0"}
	case "darwin/arm64":
		file = asset{Name: "pob-poe2-mcp-native-macos-arm64.zip", Size: 647635, Digest: "sha256:0eebe81f5221aeab6ba0f2e89e1393b8a97a9cd042c41a1fc42c1f64963dff3e"}
	default:
		return file
	}
	file.url = downloadURL(repo, tag, file.Name)
	return file
}
