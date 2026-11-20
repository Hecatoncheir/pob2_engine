package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/discovery"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/host"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/mcpapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const usage = `pob2_engine — сервис сессий Path of Exile 2

Команды:
  --version                 Версия программы и протокола
  serve                     Запустить сервер: --mode persistent|automatic
  connect                   Обнаружить сервер и удерживать аренду до EOF или Ctrl+C
  info                      Проверить готовность, протокол и возможности сервера
  session create|get|rename  Создать сессию, получить снимок или переименовать сборку
  mcp                       MCP stdio: --attach СЕССИЯ или --standalone --session СЕССИЯ
  shutdown                  Корректно завершить сервер

Подключение: --address 127.0.0.1:ПОРТ --token ТОКЕН
Или: --profile ИМЯ [--start-server] [--runtime-dir КАТАЛОГ]
Токен также читается из POB2_ENGINE_TOKEN. Справка команды: КОМАНДА --help.
Сессии пока хранятся только в памяти; игровые расчёты ещё не реализованы.
`

type UsageError struct{ Err error }

func (e *UsageError) Error() string { return e.Err.Error() }
func (e *UsageError) Unwrap() error { return e.Err }

type runner struct {
	in  io.ReadCloser
	out io.Writer
	err io.Writer
}

func Run(ctx context.Context, args []string, in io.ReadCloser, out, errOut io.Writer) error {
	r := runner{in: in, out: out, err: errOut}
	err := r.run(ctx, args)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	return err
}

func (r runner) run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return &UsageError{Err: fmt.Errorf("нужна команда; используйте --help")}
	}
	switch args[0] {
	case "--help", "-h", "help":
		_, err := fmt.Fprint(r.out, usage)
		return err
	case "--version", "version":
		return r.write(map[string]string{"version": buildinfo.Version, "protocol_version": buildinfo.ProtocolVersion})
	case "serve":
		return r.serve(ctx, args[1:])
	case "connect":
		return r.connect(ctx, args[1:])
	case "info", "shutdown":
		return r.serverCommand(ctx, args[0], args[1:])
	case "session":
		return r.session(ctx, args[1:])
	case "mcp":
		return r.mcp(ctx, args[1:])
	default:
		return &UsageError{Err: fmt.Errorf("неизвестная команда %q", args[0])}
	}
}

func (r runner) flags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(r.err)
	return fs
}

func parse(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		return &UsageError{Err: err}
	}
	if fs.NArg() != 0 {
		return &UsageError{Err: fmt.Errorf("неожиданные аргументы: %v", fs.Args())}
	}
	return nil
}

func (r runner) write(value any) error { return json.NewEncoder(r.out).Encode(value) }

type connection struct {
	address      string
	token        string
	timeout      time.Duration
	profile      string
	baseDir      string
	startServer  bool
	kind         string
	idleTimeout  time.Duration
	leaseTimeout time.Duration
}

func connectFlags(fs *flag.FlagSet) *connection {
	c := &connection{kind: "cli"}
	fs.StringVar(&c.address, "address", "", "адрес запущенного сервера 127.0.0.1:порт")
	fs.StringVar(&c.token, "token", os.Getenv("POB2_ENGINE_TOKEN"), "токен сервера; по умолчанию POB2_ENGINE_TOKEN")
	fs.DurationVar(&c.timeout, "timeout", 5*time.Second, "таймаут подключения и каждой команды")
	fs.StringVar(&c.profile, "profile", "", "профиль обнаружения сервера")
	fs.StringVar(&c.baseDir, "runtime-dir", "", "базовый каталог; профиль хранится в КАТАЛОГ/pob2-engine/ИМЯ")
	fs.BoolVar(&c.startServer, "start-server", false, "при отсутствии сервера профиля запустить automatic")
	fs.DurationVar(&c.idleTimeout, "idle-timeout", lifecycle.DefaultIdleTimeout, "задержка остановки нового automatic-сервера")
	fs.DurationVar(&c.leaseTimeout, "lease-timeout", lifecycle.DefaultLeaseTimeout, "срок аренды для нового automatic-сервера")
	return c
}

func (c *connection) dial(ctx context.Context) (*grpcapi.Client, error) {
	if c.timeout <= 0 {
		return nil, &UsageError{Err: fmt.Errorf("таймаут должен быть положительным")}
	}
	if c.profile != "" {
		if c.address != "" {
			return nil, &UsageError{Err: fmt.Errorf("выберите --profile или --address")}
		}
		return host.Connect(ctx, host.ConnectOptions{BaseDir: c.baseDir, Profile: c.profile, StartServer: c.startServer, Kind: c.kind, IdleTimeout: c.idleTimeout, LeaseTimeout: c.leaseTimeout})
	}
	if c.startServer || c.baseDir != "" {
		return nil, &UsageError{Err: fmt.Errorf("для обнаружения и запуска нужен --profile")}
	}
	return grpcapi.DialKind(ctx, c.address, c.token, c.kind)
}

type runningServer interface {
	Ready() grpcapi.Ready
	Done() <-chan struct{}
	Err() error
	Close(context.Context) error
}

func (r runner) serve(ctx context.Context, args []string) error {
	fs := r.flags("serve")
	address := fs.String("listen", "127.0.0.1:0", "локальный адрес; порт 0 выбирается автоматически")
	token := fs.String("token", os.Getenv("POB2_ENGINE_TOKEN"), "токен; без значения создаётся случайный")
	profileName := fs.String("profile", "", "опубликовать сервер в профиле")
	baseDir := fs.String("runtime-dir", "", "базовый каталог профилей")
	mode := fs.String("mode", lifecycle.Persistent, "режим сервера: persistent или automatic")
	idleTimeout := fs.Duration("idle-timeout", lifecycle.DefaultIdleTimeout, "задержка остановки без клиентов и задач")
	leaseTimeout := fs.Duration("lease-timeout", lifecycle.DefaultLeaseTimeout, "срок аренды клиента")
	if err := parse(fs, args); err != nil {
		return err
	}
	config := lifecycle.Config{Mode: *mode, IdleTimeout: *idleTimeout, LeaseTimeout: *leaseTimeout}
	server, err := startServer(*profileName, *baseDir, *address, *token, config)
	if err != nil {
		return err
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()
		_ = server.Close(stopCtx)
	}()
	if err := r.write(server.Ready()); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return nil
	case <-server.Done():
		return server.Err()
	}
}

func startServer(name, baseDir, address, token string, config lifecycle.Config) (runningServer, error) {
	if name == "" {
		if baseDir != "" {
			return nil, &UsageError{Err: fmt.Errorf("для --runtime-dir нужен --profile")}
		}
		return grpcapi.StartWithConfig(address, token, application.NewService(), config)
	}
	profile, err := discovery.Open(baseDir, name)
	if err != nil {
		return nil, err
	}
	return host.Start(profile, address, token, application.NewService(), config)
}

func (r runner) connect(ctx context.Context, args []string) error {
	fs := r.flags("connect")
	c := connectFlags(fs)
	fs.StringVar(&c.kind, "kind", "cli", "тип клиента: cli, mcp или ui")
	if err := parse(fs, args); err != nil {
		return err
	}
	if c.profile == "" && c.address == "" {
		c.profile = "default"
	}
	connectCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	client, err := c.dial(connectCtx)
	if err != nil {
		return err
	}
	defer client.Close()
	info, err := client.Info(connectCtx)
	if err != nil {
		return err
	}
	if err := r.write(info); err != nil {
		return err
	}
	cancel()
	inputDone := make(chan error, 1)
	defer r.in.Close()
	go func() { _, err := io.Copy(io.Discard, r.in); inputDone <- err }()
	select {
	case <-ctx.Done():
		return nil
	case err := <-inputDone:
		return err
	case <-client.Done():
		return fmt.Errorf("аренда сервера утрачена")
	}
}

func (r runner) serverCommand(ctx context.Context, command string, args []string) error {
	fs := r.flags(command)
	c := connectFlags(fs)
	if err := parse(fs, args); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	client, err := c.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	if command == "shutdown" {
		if err := client.Shutdown(ctx); err != nil {
			return err
		}
		return r.write(map[string]string{"state": "draining"})
	}
	info, err := client.Info(ctx)
	if err != nil {
		return err
	}
	return r.write(info)
}

func (r runner) session(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return &UsageError{Err: fmt.Errorf("нужна операция session create|get|rename")}
	}
	operation := args[0]
	if operation != "create" && operation != "get" && operation != "rename" {
		return &UsageError{Err: fmt.Errorf("неизвестная операция session %q", operation)}
	}
	fs := r.flags("session " + operation)
	c := connectFlags(fs)
	id := fs.String("session", "", "явно выбранный session_id")
	var name, requestID string
	var revision uint64
	if operation != "get" {
		fs.StringVar(&name, "name", "", "название сборки")
		fs.StringVar(&requestID, "request-id", "", "идентификатор команды для безопасного повтора")
	}
	if operation == "rename" {
		fs.Uint64Var(&revision, "revision", 0, "ожидаемая ревизия")
	}
	if err := parse(fs, args[1:]); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	client, err := c.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	var value domain.Snapshot
	switch operation {
	case "create":
		value, err = client.Create(ctx, application.CreateSession{SessionID: *id, BuildName: name, RequestID: requestID})
	case "rename":
		value, err = client.Rename(ctx, application.RenameBuild{SessionID: *id, BuildName: name, ExpectedRevision: revision, RequestID: requestID})
	default:
		value, err = client.Snapshot(ctx, *id)
	}
	if err != nil {
		return err
	}
	return r.write(value)
}

type writer struct{ io.Writer }

func (writer) Close() error { return nil }

func (r runner) mcp(ctx context.Context, args []string) error {
	fs := r.flags("mcp")
	c := connectFlags(fs)
	c.kind = "mcp"
	attach := fs.String("attach", "", "session_id на существующем сервере")
	standalone := fs.Bool("standalone", false, "создать независимую сессию до закрытия stdin")
	id := fs.String("session", "", "session_id для standalone")
	name := fs.String("name", "Новая сборка", "название сборки для standalone")
	if err := parse(fs, args); err != nil {
		return err
	}
	if c.timeout <= 0 {
		return &UsageError{Err: fmt.Errorf("таймаут должен быть положительным")}
	}
	if *standalone == (*attach != "") {
		return &UsageError{Err: fmt.Errorf("выберите один режим: --attach СЕССИЯ или --standalone --session СЕССИЯ")}
	}
	var commands application.Commands
	if *standalone {
		if c.address != "" || c.profile != "" || c.startServer || c.baseDir != "" {
			return &UsageError{Err: fmt.Errorf("--standalone не принимает параметры подключения")}
		}
		service := application.NewService()
		if _, err := service.Create(ctx, application.CreateSession{SessionID: *id, BuildName: *name, RequestID: "standalone-init"}); err != nil {
			return err
		}
		commands = service
	} else {
		if *id != "" {
			return &UsageError{Err: fmt.Errorf("при --attach укажите сессию только в --attach")}
		}
		connectCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()
		client, err := c.dial(connectCtx)
		if err != nil {
			return err
		}
		defer client.Close()
		if _, err := client.Snapshot(connectCtx, *attach); err != nil {
			return err
		}
		*id, commands = *attach, client
	}
	return mcpapi.New(commands, *id, c.timeout).Run(ctx, &mcp.IOTransport{Reader: r.in, Writer: writer{r.out}})
}
