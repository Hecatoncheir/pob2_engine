package pobwindow

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBridgeInstallAndRestoreOriginalFile(t *testing.T) {
	t.Setenv("WINEPREFIX", t.TempDir())
	state, installation := t.TempDir(), t.TempDir()
	source := []byte("-- Оригинальный файл\r\nlocal launch = {}\r\n")
	launch := prepareWindowInstallation(t, installation, source)
	installed, err := InstallBridge(state, installation, false)
	if err != nil || !installed.RestartRequired {
		t.Fatalf("установка: %+v, %v", installed, err)
	}
	if err := CheckBridge(state, installation); err != nil {
		t.Fatal(err)
	}
	assertLaunchBytes(t, installed.Backup, source)

	existing, err := InstallBridge(state, installation, true)
	if err != nil || existing.RestartRequired || existing.Backup != installed.Backup {
		t.Fatalf("повторная установка: %+v, %v", existing, err)
	}
	removed, err := UninstallBridge(state, installation)
	if err != nil || !removed.RestartRequired {
		t.Fatalf("снятие моста: %+v, %v", removed, err)
	}
	assertLaunchBytes(t, launch, source)
}

func prepareWindowInstallation(t *testing.T, directory string, launch []byte) string {
	t.Helper()
	for name, content := range map[string][]byte{
		"Launch.lua": launch, "GameVersions.lua": {}, "Path of Building-PoE2.exe": {},
	} {
		if err := os.WriteFile(filepath.Join(directory, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(directory, "Launch.lua")
}

func assertLaunchBytes(t *testing.T, file string, expected []byte) {
	t.Helper()
	content, err := os.ReadFile(file)
	if err != nil || !bytes.Equal(content, expected) {
		t.Fatalf("Launch.lua: %q, %v", content, err)
	}
}
