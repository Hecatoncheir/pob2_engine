package pob

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/privatefs"
)

const (
	repository     = "PathOfBuildingCommunity/PathOfBuilding-PoE2"
	latestURL      = "https://api.github.com/repos/" + repository + "/releases/latest"
	InstallTimeout = 5 * time.Minute
)

var tags = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
var digestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var commitPattern = regexp.MustCompile(`^[a-f0-9]{40}$`)

type Installation struct {
	Version         string `json:"version"`
	SHA256          string `json:"sha256"`
	Commit          string `json:"commit"`
	Directory       string `json:"-"`
	NativeDirectory string `json:"-"`
}

type State struct {
	Status           string `json:"status"`
	InstalledVersion string `json:"installed_version,omitempty"`
	LatestVersion    string `json:"latest_version,omitempty"`
	ReleaseURL       string `json:"release_url,omitempty"`
	Message          string `json:"message,omitempty"`
	RestartRequired  bool   `json:"restart_required,omitempty"`
}

type Manager struct {
	root, local, localNative string
	client                   *http.Client
	probe                    func(context.Context, Installation) error
	nativeRuntime            asset
}

func Local() (*Manager, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	return &Manager{root: filepath.Join(cache, "pob2-engine", "pob"),
		local: os.Getenv("POB2_ENGINE_POB_DIR"), localNative: os.Getenv("POB2_ENGINE_POB_NATIVE_DIR"),
		client: &http.Client{Timeout: InstallTimeout, CheckRedirect: checkRedirect}, probe: probeInstallation,
		nativeRuntime: nativeAsset(runtime.GOOS, runtime.GOARCH)}, nil
}

func (m *Manager) Installed() (Installation, error) {
	if m.local != "" {
		return localInstallation(m.local, m.localNative)
	}
	var i Installation
	data, err := os.ReadFile(filepath.Join(m.root, "current.json"))
	if err != nil {
		return i, err
	}
	if err := json.Unmarshal(data, &i); err != nil {
		return i, err
	}
	if !tags.MatchString(i.Version) || !digestPattern.MatchString(i.SHA256) || !commitPattern.MatchString(i.Commit) {
		return i, fmt.Errorf("повреждён манифест установки PoB")
	}
	return m.locate(i), nil
}

func localInstallation(dir, native string) (Installation, error) {
	directory, err := filepath.Abs(dir)
	if err != nil {
		return Installation{}, err
	}
	if native != "" {
		native, err = filepath.Abs(native)
		if err != nil {
			return Installation{}, err
		}
	}
	version := readVersion(directory)
	return Installation{Version: version, Directory: directory, NativeDirectory: native}, nil
}

func readVersion(directory string) string {
	for _, dir := range []string{directory, filepath.Join(directory, "src")} {
		data, err := os.ReadFile(filepath.Join(dir, "manifest.xml"))
		if err != nil {
			continue
		}
		var manifest struct {
			Version struct {
				Number string `xml:"number,attr"`
			} `xml:"Version"`
		}
		if xml.Unmarshal(data, &manifest) == nil && tags.MatchString("v"+manifest.Version.Number) {
			return "v" + manifest.Version.Number
		}
	}
	return "local"
}

func (m *Manager) locate(i Installation) Installation {
	i.Directory = filepath.Join(m.root, i.Version+"-"+i.SHA256[:12])
	if regularFile(filepath.Join(i.Directory, "native", "lua-utf8.so")) {
		i.NativeDirectory = filepath.Join(i.Directory, "native")
	}
	return i
}

// Ensure устанавливает PoB при первом запуске. Уже установленная версия работает офлайн.
func (m *Manager) Ensure(ctx context.Context) (Installation, error) {
	i, err := m.Installed()
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		return i, err
	}
	if _, err := m.installUpdate(ctx, initialRelease(), State{}); err != nil {
		return Installation{}, err
	}
	return m.Installed()
}

func (m *Manager) Update(ctx context.Context, apply bool) (State, error) {
	installed, err := m.Installed()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return State{Status: "invalid_installation"}, err
	}
	if err == nil {
		if _, err := installed.Layout(); err != nil {
			return State{Status: "invalid_installation", InstalledVersion: installed.Version}, err
		}
	}
	state := State{Status: "up_to_date", InstalledVersion: installed.Version}
	release, err := m.latest(ctx)
	if err != nil {
		state.Status = "unavailable"
		return state, err
	}
	state.LatestVersion, state.ReleaseURL = release.Tag, "https://github.com/"+repository+"/releases/tag/"+release.Tag
	if m.local != "" {
		state.Status = "local_installation"
		return state, nil
	}
	if installed.Version != "" && !newer(release.Tag, installed.Version) {
		return state, nil
	}
	state.Status = "available"
	if apply {
		return m.installUpdate(ctx, release, state)
	}
	return state, nil
}

func newer(next, installed string) bool {
	a, b := strings.Split(next[1:], "."), strings.Split(installed[1:], ".")
	for index, part := range a {
		left, _ := strconv.ParseUint(part, 10, 64)
		right, _ := strconv.ParseUint(b[index], 10, 64)
		if left != right {
			return left > right
		}
	}
	return false
}

func (m *Manager) installUpdate(ctx context.Context, release release, state State) (State, error) {
	if err := os.MkdirAll(m.root, 0700); err != nil {
		return state, err
	}
	lock, err := waitForInstaller(ctx, filepath.Join(m.root, "install.lock"))
	if err != nil {
		return state, fmt.Errorf("блокировка установки PoB: %w", err)
	}
	defer lock.Close()
	current, err := m.Installed()
	if err == nil && !newer(release.Tag, current.Version) {
		state.Status, state.InstalledVersion = "up_to_date", current.Version
		return state, nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return state, err
	}
	if err := m.install(ctx, release); err != nil {
		return state, err
	}
	state.Status, state.InstalledVersion = "updated", release.Tag
	state.RestartRequired = current.Version != ""
	return state, nil
}

func waitForInstaller(ctx context.Context, path string) (*os.File, error) {
	wait := time.NewTicker(100 * time.Millisecond)
	defer wait.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		file, err := privatefs.TryLock(path)
		if !errors.Is(err, privatefs.ErrLocked) {
			return file, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-wait.C:
		}
	}
}

func probeInstallation(ctx context.Context, i Installation) error {
	c, err := NewCalculator(i)
	if err != nil {
		return err
	}
	defer c.Close()
	return c.start(ctx)
}
