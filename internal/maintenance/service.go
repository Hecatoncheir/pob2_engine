package maintenance

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
)

type Commands interface {
	Update(ctx context.Context, apply bool, targets ...string) (UpdateResult, error)
	Doctor(ctx context.Context) (Report, error)
}

type Service struct {
	version string
	client  *http.Client
	goos    string
	goarch  string
	path    string
	image   os.FileInfo
	pathErr error
}

// Local запоминает путь и исходный файл один раз на процесс.
var Local = sync.OnceValue(func() *Service {
	s := &Service{version: buildinfo.Version, client: githubClient(), goos: runtime.GOOS, goarch: runtime.GOARCH}
	s.path, s.image, s.pathErr = installedExecutable()
	return s
})

func installedExecutable() (string, os.FileInfo, error) {
	name, err := os.Executable()
	if err != nil {
		return "", nil, err
	}
	path, err := filepath.EvalSymlinks(name)
	if err != nil {
		return name, nil, err
	}
	image, err := snapshotFile(path)
	return path, image, err
}

// На Windows Stat по открытому файлу фиксирует идентификатор до замены пути.
func snapshotFile(path string) (os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return file.Stat()
}
