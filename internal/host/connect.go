package host

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/discovery"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/storage"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ConnectOptions struct {
	BaseDir         string
	DataDir         string
	Profile         string
	StartServer     bool
	Executable      string
	Kind            string
	IdleTimeout     time.Duration
	LeaseTimeout    time.Duration
	AllowMemory     bool
	expectedStorage string
}

func Connect(ctx context.Context, options ConnectOptions) (*grpcapi.Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if options.Kind == "" {
		options.Kind = "cli"
	}
	if options.DataDir != "" {
		var err error
		options.expectedStorage, err = storage.Path(options.DataDir, options.Profile)
		if err != nil {
			return nil, err
		}
	}
	profile, err := discovery.Open(options.BaseDir, options.Profile)
	if err != nil {
		return nil, err
	}
	client, _, err := tryConnect(ctx, profile, options)
	if err == nil || !options.StartServer {
		return client, err
	}
	lock, err := profile.Lock(ctx, "launch")
	if err != nil {
		return nil, err
	}
	defer lock.Close()
	return ensure(ctx, profile, options)
}

func tryConnect(ctx context.Context, profile discovery.Profile, options ConnectOptions) (*grpcapi.Client, bool, error) {
	record, err := profile.Read()
	if err != nil {
		return nil, false, err
	}
	probeCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	info, err := grpcapi.Inspect(probeCtx, record.Address, record.Token)
	if err != nil {
		return nil, status.Code(err) == codes.Unauthenticated, err
	}
	if record.ProtocolVersion != buildinfo.ProtocolVersion || info.GetProtocolVersion() != buildinfo.ProtocolVersion {
		return nil, true, fmt.Errorf("несовместимый протокол сервера профиля: %q", info.GetProtocolVersion())
	}
	if record.InstanceID != info.GetInstanceId() {
		return nil, true, fmt.Errorf("идентификатор экземпляра не совпадает с файлом подключения")
	}
	if info.GetState() != "ready" {
		return nil, true, lifecycle.ErrDraining
	}
	if !options.AllowMemory && (info.GetStorageMode() != "disk" || info.GetStoragePath() == "") {
		return nil, true, fmt.Errorf("сервер профиля не поддерживает сохранение; завершите его и запустите новую версию")
	}
	if options.expectedStorage != "" && !samePath(options.expectedStorage, info.GetStoragePath()) {
		return nil, true, fmt.Errorf("сервер профиля использует другой каталог данных: %s", info.GetStoragePath())
	}
	client, err := grpcapi.DialKind(ctx, record.Address, record.Token, options.Kind)
	if err != nil && status.Code(err) == codes.Unavailable {
		return nil, true, lifecycle.ErrDraining
	}
	if err == nil && client.InstanceID() != record.InstanceID {
		client.Close()
		return nil, true, fmt.Errorf("идентификатор экземпляра изменился при подключении")
	}
	return client, true, err
}

func samePath(first, second string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(first), filepath.Clean(second))
	}
	return filepath.Clean(first) == filepath.Clean(second)
}

func ensure(ctx context.Context, profile discovery.Profile, options ConnectOptions) (*grpcapi.Client, error) {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	var process *child
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		client, alive, err := tryConnect(ctx, profile, options)
		if err == nil {
			return client, nil
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if process == nil {
			var startErr error
			process, startErr = startIfUnlocked(profile, options)
			if startErr != nil {
				return nil, startErr
			}
			if process == nil && alive && !errors.Is(err, lifecycle.ErrDraining) {
				return nil, err
			}
		}
		if process != nil {
			select {
			case <-process.done:
				return nil, process.failure(profile)
			default:
			}
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("сервер профиля не готов: %w (последняя проверка: %v)", ctx.Err(), err)
		case <-ticker.C:
		}
	}
}

func startIfUnlocked(profile discovery.Profile, options ConnectOptions) (*child, error) {
	lock, err := profile.TryLock("server")
	if errors.Is(err, discovery.ErrLocked) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := lock.Close(); err != nil {
		return nil, err
	}
	return launch(profile, options)
}

type child struct {
	done chan struct{}
	err  error
}

func launch(profile discovery.Profile, options ConnectOptions) (*child, error) {
	config, err := (lifecycle.Config{Mode: lifecycle.Automatic, IdleTimeout: options.IdleTimeout, LeaseTimeout: options.LeaseTimeout}).Normalize()
	if err != nil {
		return nil, err
	}
	executable := options.Executable
	if executable == "" {
		executable, err = os.Executable()
		if err != nil {
			return nil, err
		}
	}
	args := []string{"serve", "--profile", options.Profile, "--mode", lifecycle.Automatic, "--idle-timeout", config.IdleTimeout.String(), "--lease-timeout", config.LeaseTimeout.String()}
	if options.BaseDir != "" {
		args = append(args, "--runtime-dir", options.BaseDir)
	}
	if options.DataDir != "" {
		args = append(args, "--data-dir", options.DataDir)
	}
	log, err := os.OpenFile(filepath.Join(profile.Dir, "server.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer log.Close()
	cmd := exec.Command(executable, args...)
	cmd.Env = append(os.Environ(), "POB2_ENGINE_TOKEN=")
	cmd.Stderr = log
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	process := &child{done: make(chan struct{})}
	go func() { process.err = cmd.Wait(); close(process.done) }()
	return process, nil
}

func (c *child) failure(profile discovery.Profile) error {
	data, _ := os.ReadFile(filepath.Join(profile.Dir, "server.log"))
	if len(data) > 2048 {
		data = data[len(data)-2048:]
	}
	return fmt.Errorf("сервер завершился при запуске: %v; %s", c.err, data)
}
