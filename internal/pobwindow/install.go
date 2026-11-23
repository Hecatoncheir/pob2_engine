package pobwindow

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Hecatoncheir/pob2_engine/internal/poblegacy"
	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
)

const (
	beginHook = "-- BEGIN POB_POE2_MCP_BRIDGE"
	endHook   = "-- END POB_POE2_MCP_BRIDGE"
)

type bridgeConfig struct {
	Token           string `json:"token"`
	InstallDir      string `json:"installDir"`
	Backup          string `json:"backup"`
	CleanSourceHash string `json:"cleanSourceHash"`
	HostPlatform    string `json:"hostPlatform,omitempty"`
}

type bridgeInstallation struct {
	EngineDirectory string `json:"engineDir"`
	Backup          string `json:"backup,omitempty"`
	RestartRequired bool   `json:"restartRequired"`
}

type bridgeSource struct {
	root, launch    string
	original, clean []byte
	mode            os.FileMode
}

func openBridgeSource(directory string) (bridgeSource, error) {
	root, err := filepath.Abs(directory)
	if err != nil {
		return bridgeSource{}, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return bridgeSource{}, err
	}
	launch, err := launchFile(root)
	if err != nil {
		return bridgeSource{}, err
	}
	info, err := os.Lstat(launch)
	if err != nil {
		return bridgeSource{}, err
	}
	original, err := os.ReadFile(launch)
	if err != nil {
		return bridgeSource{}, err
	}
	clean, err := removeHook(original)
	return bridgeSource{root: root, launch: launch, original: original, clean: clean, mode: info.Mode().Perm()}, err
}

func launchFile(root string) (string, error) {
	if !hasExecutable(root) {
		return "", &BridgeError{"ENGINE_NOT_CONFIGURED", "Выберите каталог оригинального приложения Path of Building для PoE 2"}
	}
	for _, folder := range []string{root, filepath.Join(root, "src")} {
		launch := filepath.Join(folder, "Launch.lua")
		info, err := os.Lstat(launch)
		if err == nil && info.Mode().IsRegular() && isFile(filepath.Join(folder, "GameVersions.lua")) {
			return launch, nil
		}
	}
	return "", &BridgeError{"ENGINE_NOT_CONFIGURED", "Не найден штатный Launch.lua Path of Building для PoE 2"}
}

func hasExecutable(root string) bool {
	for _, name := range []string{"Path of Building-PoE2.exe", "runtime/Path of Building-PoE2.exe", "runtime/Path{space}of{space}Building-PoE2.exe"} {
		if isFile(filepath.Join(root, filepath.FromSlash(name))) {
			return true
		}
	}
	return false
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func removeHook(source []byte) ([]byte, error) {
	start, finish := bytes.Index(source, []byte(beginHook)), bytes.Index(source, []byte(endHook))
	if start < 0 && finish < 0 {
		return source, nil
	}
	if start < 0 || finish < start || bytes.Count(source, []byte(beginHook)) != 1 || bytes.Count(source, []byte(endHook)) != 1 {
		return nil, &BridgeError{"INVALID_BRIDGE_HOOK", "Блок моста в Launch.lua повреждён"}
	}
	clean := append([]byte(nil), bytes.TrimRight(source[:start], "\r\n")...)
	return append(clean, source[finish+len(endHook):]...), nil
}

func InstallBridge(state, directory string, ensure bool) (bridgeInstallation, error) {
	return withBridgeLock(state, func(state string) (bridgeInstallation, error) {
		return installLocked(state, directory, ensure)
	})
}

func withBridgeLock(state string, action func(string) (bridgeInstallation, error)) (bridgeInstallation, error) {
	state, err := filepath.Abs(state)
	if err != nil {
		return bridgeInstallation{}, err
	}
	if err := privatefs.EnsureDir(state); err != nil {
		return bridgeInstallation{}, err
	}
	lock, err := privatefs.TryLock(filepath.Join(state, "install.lock"))
	if err != nil {
		return bridgeInstallation{}, err
	}
	defer lock.Close()
	return action(state)
}

func installLocked(state, directory string, ensure bool) (bridgeInstallation, error) {
	source, err := openBridgeSource(directory)
	if err != nil {
		return bridgeInstallation{}, err
	}
	config, original, err := loadBridgeConfig(state)
	if err != nil {
		return bridgeInstallation{}, err
	}
	if ensure && configured(source, state, config) {
		return bridgeInstallation{source.root, config.Backup, false}, nil
	}
	return writeBridge(state, source, config, original)
}

func loadBridgeConfig(state string) (bridgeConfig, []byte, error) {
	var config bridgeConfig
	content, err := os.ReadFile(filepath.Join(state, "config.json"))
	if errors.Is(err, os.ErrNotExist) {
		return config, nil, nil
	}
	if err != nil {
		return config, nil, err
	}
	if err := json.Unmarshal(content, &config); err != nil {
		return config, nil, fmt.Errorf("повреждена конфигурация моста PoB: %w", err)
	}
	return config, content, nil
}

func configured(source bridgeSource, state string, config bridgeConfig) bool {
	path, err := bridgeScriptPath(state, source.root)
	return err == nil && bytes.HasSuffix(source.original, bytes.TrimLeft(bridgeHook(path), "\n")) && len(config.Token) >= 32 &&
		sameDirectory(source.root, config.InstallDir) && config.CleanSourceHash == cleanSourceHash(source.clean) &&
		isFile(filepath.Join(state, "lua", "gui_bridge.lua"))
}

func CheckBridge(state, directory string) error {
	source, err := openBridgeSource(directory)
	if err != nil {
		return err
	}
	config, _, err := loadBridgeConfig(state)
	if err != nil {
		return err
	}
	if !configured(source, state, config) {
		return &BridgeError{"BRIDGE_NOT_CONFIGURED", "Мост PoB нужно установить или обновить"}
	}
	return nil
}

func writeBridge(state string, source bridgeSource, old bridgeConfig, originalConfig []byte) (bridgeInstallation, error) {
	bridgePath, err := bridgeScriptPath(state, source.root)
	if err != nil {
		return bridgeInstallation{}, err
	}
	backup, err := saveBackup(state, source, old)
	if err != nil {
		return bridgeInstallation{}, err
	}
	if err := poblegacy.Unpack(state, "lua"); err != nil {
		return bridgeInstallation{}, err
	}
	token, err := newToken()
	if err != nil {
		return bridgeInstallation{}, err
	}
	config := bridgeConfig{Token: token, InstallDir: source.root, Backup: backup, CleanSourceHash: cleanSourceHash(source.clean)}
	if runtime.GOOS == "linux" {
		config.HostPlatform = "linux"
	}
	if err := activateBridge(state, source, bridgePath, config, originalConfig); err != nil {
		return bridgeInstallation{}, err
	}
	return bridgeInstallation{source.root, backup, true}, nil
}

func newToken() (string, error) {
	var data [32]byte
	_, err := rand.Read(data[:])
	return hex.EncodeToString(data[:]), err
}

func bridgeScriptPath(state, installation string) (string, error) {
	location := filepath.ToSlash(state)
	if runtime.GOOS == "linux" {
		prefix, err := WinePrefix(installation, "")
		if err != nil {
			return "", err
		}
		location, err = pathForWine(state, prefix)
		if err != nil {
			return "", err
		}
	}
	if strings.Contains(location, "]]") {
		return "", &BridgeError{"INVALID_PATH", "Путь к каталогу моста содержит ]]"}
	}
	return location, nil
}

func saveBackup(state string, source bridgeSource, old bridgeConfig) (string, error) {
	folder := filepath.Join(state, "backups")
	if err := privatefs.EnsureDir(folder); err != nil {
		return "", err
	}
	if old.CleanSourceHash == cleanSourceHash(source.clean) && filepath.Dir(old.Backup) == folder && isFile(old.Backup) {
		return old.Backup, nil
	}
	content := source.clean
	digest := sha256.Sum256(content)
	name := filepath.Join(folder, fmt.Sprintf("Launch-%x.lua", digest))
	if !isFile(name) {
		return name, replaceFile(name, content, 0600)
	}
	return name, nil
}

func cleanSourceHash(source []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(bytes.TrimSpace(source)))
}

func activateBridge(state string, source bridgeSource, path string, config bridgeConfig, old []byte) error {
	content, err := json.Marshal(config)
	if err != nil {
		return err
	}
	configPath := filepath.Join(state, "config.json")
	if err := replaceFile(configPath, content, 0600); err != nil {
		return err
	}
	launch := append(bytes.TrimRight(source.clean, "\r\n"), bridgeHook(path)...)
	if err := replaceFile(source.launch, launch, source.mode); err != nil {
		return errors.Join(err, restoreConfig(configPath, old))
	}
	return nil
}

func restoreConfig(path string, old []byte) error {
	if old == nil {
		return os.Remove(path)
	}
	return replaceFile(path, old, 0600)
}

func bridgeHook(path string) []byte {
	return fmt.Appendf(nil, "\n\n%s\nif not POB_MCP_HEADLESS then\n    local ok, message = pcall(function()\n        dofile([[%s/lua/gui_bridge.lua]])(launch, [[%s]])\n    end)\n    if not ok then ConPrintf(\"MCP: %%s\", tostring(message)) end\nend\n%s\n", beginHook, path, path, endHook)
}

func replaceFile(path string, content []byte, mode os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".pob2-bridge-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err := writeFile(file, content, mode); err != nil {
		return err
	}
	return privatefs.Replace(file.Name(), path)
}

func writeFile(file *os.File, content []byte, mode os.FileMode) (err error) {
	defer func() { err = errors.Join(err, file.Close()) }()
	if err := file.Chmod(mode); err != nil {
		return err
	}
	if _, err := file.Write(content); err != nil {
		return err
	}
	return file.Sync()
}

func UninstallBridge(state, directory string) (bridgeInstallation, error) {
	return withBridgeLock(state, func(state string) (bridgeInstallation, error) {
		return uninstallLocked(state, directory)
	})
}

func uninstallLocked(state, directory string) (bridgeInstallation, error) {
	source, err := openBridgeSource(directory)
	if err != nil {
		return bridgeInstallation{}, err
	}
	result := bridgeInstallation{EngineDirectory: source.root, RestartRequired: !bytes.Equal(source.original, source.clean)}
	if !result.RestartRequired {
		return result, nil
	}
	config, _, err := loadBridgeConfig(state)
	if err != nil {
		return result, err
	}
	content := originalLaunch(state, source, config)
	return result, replaceFile(source.launch, content, source.mode)
}

func originalLaunch(state string, source bridgeSource, config bridgeConfig) []byte {
	if config.CleanSourceHash != cleanSourceHash(source.clean) || filepath.Dir(config.Backup) != filepath.Join(state, "backups") {
		return source.clean
	}
	backup, err := os.ReadFile(config.Backup)
	if err != nil || cleanSourceHash(backup) != config.CleanSourceHash {
		return source.clean
	}
	return backup
}
