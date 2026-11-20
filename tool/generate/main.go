// Генератор запускается из корня: go run ./tool/generate [-check].
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	check := flag.Bool("check", false, "сравнить результат с сохранённым API")
	flag.Parse()
	if err := generate(*check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate(check bool) error {
	dart, err := prepareTools()
	if err != nil {
		return err
	}
	if err := run(dart, "pub", "-C", "tool/proto", "get", "--enforce-lockfile"); err != nil {
		return err
	}
	if !check {
		return run("buf", "generate")
	}
	return checkGenerated()
}

func prepareTools() (string, error) {
	// Flutter-обёртка может писать в stdout; плагину нужен чистый двоичный поток.
	dart, err := exec.LookPath("dart")
	if err != nil {
		return "", err
	}
	if strings.EqualFold(filepath.Ext(dart), ".bat") || strings.EqualFold(filepath.Ext(dart), ".cmd") {
		dart = filepath.Join(filepath.Dir(dart), "cache", "dart-sdk", "bin", "dart.exe")
		if _, err := os.Stat(dart); err != nil {
			return "", fmt.Errorf("не найден Dart SDK Flutter: %w", err)
		}
	}
	if err := os.Setenv("PATH", filepath.Dir(dart)+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		return "", err
	}
	version, err := exec.Command("buf", "--version").Output()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(version)) != "1.73.0" {
		return "", fmt.Errorf("для генерации нужен Buf 1.73.0")
	}
	return dart, nil
}

func checkGenerated() error {
	temp, err := os.MkdirTemp("", "pob2-proto-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	if err := run("buf", "generate", "--output", temp); err != nil {
		return err
	}
	for _, dir := range []string{"gen/go", "gen/dart/lib"} {
		if err := compare(filepath.Join(temp, dir), dir); err != nil {
			return err
		}
	}
	fmt.Println("Go- и Dart-API соответствуют контракту.")
	return nil
}

func compare(expected, current string) error {
	want, err := files(expected)
	if err != nil {
		return err
	}
	got, err := files(current)
	if err != nil {
		return err
	}
	if len(want) != len(got) {
		return fmt.Errorf("набор файлов %s отличается: выполните генерацию", current)
	}
	for name, data := range want {
		if !bytes.Equal(data, got[name]) {
			return fmt.Errorf("устарел %s: выполните генерацию", filepath.Join(current, name))
		}
	}
	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func files(root string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result[name] = data
		return nil
	})
	return result, err
}
