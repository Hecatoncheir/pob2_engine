package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/discovery"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"github.com/Hecatoncheir/pob2_engine/internal/host"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"github.com/Hecatoncheir/pob2_engine/internal/storage"
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
  update                    Установить стабильный выпуск или проверить его наличие с --check
  doctor                    Проверить бинарник, набор данных и состояние выбранного сервера
  session create|get|rename  Создать сессию, получить снимок или переименовать сборку
  build import|edit|export|compare  Операции полной сборки выбранной сессии
  build-report               Прочитать подробный отчёт Lua о сборке
  equipment-details          Прочитать описание и сравнить предметы
  tree-geometry              Получить узлы и связи активного дерева
  tree-node-tooltip          Прочитать подсказку и путь узла дерева
  tree-power                 Найти влияние узлов на показатели сборки
  tree-texture               Прочитать текстуру дерева в формате PNG
  calculate                 Рассчитать --file XML либо выбранную --session
  catalog                   Поиск данных текущего набора: --kind gems|classes|items|passives
  preview-support           Сравнить поддержку в --session без записи
  pob-window                Прочитать открытую сборку и данные оригинального окна PoB
  pob-data                  Проверить, установить, обновить или откатить данные PoB
  storage backup|verify|restore  Резервные копии остановленного профиля
  mcp                       MCP stdio: --attach СЕССИЯ или --standalone --session СЕССИЯ
  workbench [--grpc]        Рабочая копия XML через MCP stdio или локальный gRPC
  editor                    MCP stdio для разрешённого окна PoB II: --endpoint ФАЙЛ
  shutdown                  Корректно завершить сервер

Подключение: --address 127.0.0.1:ПОРТ --token ТОКЕН
Или: --profile ИМЯ [--start-server] [--runtime-dir КАТАЛОГ] [--data-dir КАТАЛОГ]
Токен также читается из POB2_ENGINE_TOKEN. Справка команды: КОМАНДА --help.
Сессии профиля сохраняются на диск. --ephemeral и MCP --standalone используют память.
Каталог и расчёты по умолчанию предоставляет оригинальный PoB на LuaJIT.
Прежний режим Go ограничен Fireball/Frostbolt/Spark. Проверяйте status, issues и coverage.
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
	case "update", "doctor":
		return r.maintenance(ctx, args[0], args[1:])
	case "session":
		return r.session(ctx, args[1:])
	case "mcp":
		return r.mcp(ctx, args[1:])
	case "workbench":
		return r.workbench(ctx, args[1:])
	case "editor":
		return r.localEditor(ctx, args[1:])
	case "build":
		return r.build(ctx, args[1:])
	case "build-report":
		return r.buildReport(ctx, args[1:])
	case "equipment-details":
		return r.equipmentDetails(ctx, args[1:])
	case "tree-geometry":
		return r.treeGeometry(ctx, args[1:])
	case "tree-node-tooltip":
		return r.treeNodeTooltip(ctx, args[1:])
	case "tree-power":
		return r.treePower(ctx, args[1:])
	case "tree-texture":
		return r.treeTexture(ctx, args[1:])
	case "calculate":
		return r.calculate(ctx, args[1:])
	case "catalog":
		return r.catalog(ctx, args[1:])
	case "preview-support":
		return r.previewSupport(ctx, args[1:])
	case "pob-window":
		return r.pobWindow(ctx, args[1:])
	case "pob-window-control":
		return r.pobWindowControl(args[1:])
	case "pob-data":
		return r.pobData(ctx, args[1:])
	case "storage":
		return r.storage(ctx, args[1:])
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
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return &UsageError{Err: fmt.Errorf("неожиданные аргументы: %v", fs.Args())}
	}
	return nil
}

func parseFlags(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		return &UsageError{Err: err}
	}
	return nil
}

func (r runner) write(value any) error {
	if snapshot, ok := value.(domain.Snapshot); ok {
		view, err := application.ViewSnapshot(snapshot)
		if err != nil {
			return err
		}
		value = view
	}
	return json.NewEncoder(r.out).Encode(value)
}

type connection struct {
	address      string
	token        string
	timeout      time.Duration
	profile      string
	baseDir      string
	dataDir      string
	startServer  bool
	kind         string
	idleTimeout  time.Duration
	leaseTimeout time.Duration
	allowMemory  bool
}

func connectFlags(fs *flag.FlagSet) *connection {
	return connectFlagsWithTimeout(fs, 5*time.Second)
}

func connectFlagsWithTimeout(fs *flag.FlagSet, timeout time.Duration) *connection {
	c := &connection{kind: "cli"}
	fs.StringVar(&c.address, "address", "", "адрес запущенного сервера 127.0.0.1:порт")
	fs.StringVar(&c.token, "token", os.Getenv("POB2_ENGINE_TOKEN"), "токен сервера; по умолчанию POB2_ENGINE_TOKEN")
	fs.DurationVar(&c.timeout, "timeout", timeout, "таймаут подключения и каждой команды")
	fs.StringVar(&c.profile, "profile", "", "профиль обнаружения сервера")
	fs.StringVar(&c.baseDir, "runtime-dir", "", "базовый каталог; профиль хранится в КАТАЛОГ/pob2-engine/ИМЯ")
	fs.StringVar(&c.dataDir, "data-dir", "", "базовый каталог данных; по умолчанию os.UserConfigDir()")
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
		return host.Connect(ctx, host.ConnectOptions{BaseDir: c.baseDir, DataDir: c.dataDir, Profile: c.profile, StartServer: c.startServer, Kind: c.kind, IdleTimeout: c.idleTimeout, LeaseTimeout: c.leaseTimeout, AllowMemory: c.allowMemory})
	}
	if c.startServer || c.baseDir != "" || c.dataDir != "" {
		return nil, &UsageError{Err: fmt.Errorf("для обнаружения и запуска нужен --profile")}
	}
	return grpcapi.DialKind(ctx, c.address, c.token, c.kind)
}

type runningServer interface {
	Ready() grpcapi.Ready
	Done() <-chan struct{}
	Draining() <-chan struct{}
	Err() error
	Close(context.Context) error
}

func (r runner) serve(ctx context.Context, args []string) error {
	fs := r.flags("serve")
	address := fs.String("listen", "127.0.0.1:0", "локальный адрес; порт 0 выбирается автоматически")
	token := fs.String("token", os.Getenv("POB2_ENGINE_TOKEN"), "токен; без значения создаётся случайный")
	profileName := fs.String("profile", "", "опубликовать сервер в профиле")
	baseDir := fs.String("runtime-dir", "", "базовый каталог профилей")
	dataDir := fs.String("data-dir", "", "базовый каталог сохранений; по умолчанию os.UserConfigDir()")
	ephemeral := fs.Bool("ephemeral", false, "работать только в памяти без профиля и сохранения")
	mode := fs.String("mode", lifecycle.Persistent, "режим сервера: persistent или automatic")
	idleTimeout := fs.Duration("idle-timeout", lifecycle.DefaultIdleTimeout, "задержка остановки без клиентов и задач")
	leaseTimeout := fs.Duration("lease-timeout", lifecycle.DefaultLeaseTimeout, "срок аренды клиента")
	if err := parse(fs, args); err != nil {
		return err
	}
	config := lifecycle.Config{Mode: *mode, IdleTimeout: *idleTimeout, LeaseTimeout: *leaseTimeout}
	if *ephemeral && (*profileName != "" || *baseDir != "" || *dataDir != "") {
		return &UsageError{Err: fmt.Errorf("--ephemeral не принимает профиль и каталоги хранения")}
	}
	if !*ephemeral && *profileName == "" {
		*profileName = "default"
	}
	if err := grpcapi.ValidateAddress(*address); err != nil {
		return err
	}
	calculator, closeCalculator, err := openCalculator(ctx)
	if err != nil {
		return err
	}
	defer closeCalculator()
	server, err := startServer(ctx, *profileName, *baseDir, *dataDir, *address, *token, config, calculator)
	if err != nil {
		return err
	}
	return r.waitForServer(ctx, server)
}

func (r runner) waitForServer(ctx context.Context, server runningServer) (resultErr error) {
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()
		if closeErr := server.Close(stopCtx); closeErr != nil && !errors.Is(resultErr, closeErr) {
			resultErr = errors.Join(resultErr, closeErr)
		}
	}()
	if err := r.write(server.Ready()); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return nil
	case <-server.Draining():
		return nil
	case <-server.Done():
		return server.Err()
	}
}

func startServer(ctx context.Context, name, baseDir, dataDir, address, token string, config lifecycle.Config, calculator *pob.Calculator) (runningServer, error) {
	if name == "" {
		if baseDir != "" {
			return nil, &UsageError{Err: fmt.Errorf("для --runtime-dir нужен --profile")}
		}
		return grpcapi.StartWithConfig(address, token, application.NewServiceWithCalculator(calculator), config)
	}
	profile, err := discovery.Open(baseDir, name)
	if err != nil {
		return nil, err
	}
	path, err := storage.Path(dataDir, profile.Name)
	if err != nil {
		return nil, err
	}
	calculator.EnableTreeSceneCache(filepath.Dir(path))
	return host.StartDurableWithCalculator(ctx, profile, address, token, dataDir, config, calculator)
}

func (r runner) connect(ctx context.Context, args []string) error {
	fs := r.flags("connect")
	c := connectFlags(fs)
	fs.StringVar(&c.kind, "kind", "cli", "тип клиента: cli, mcp или ui")
	connectionJSON := fs.Bool("connection-json", false, "вывести адрес, токен и аренду для дочернего клиента; не сохранять вывод в журнал")
	if err := parse(fs, args); err != nil {
		return err
	}
	if c.profile == "" && c.address == "" {
		c.profile = "default"
	}
	connectCtx, cancel := context.WithTimeout(ctx, c.connectTimeout())
	defer cancel()
	client, err := c.dial(connectCtx)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := r.writeConnection(connectCtx, client, c, *connectionJSON); err != nil {
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

func (c *connection) usesServer() bool {
	return c.address != "" || c.profile != "" || c.startServer || c.baseDir != "" || c.dataDir != ""
}

func (c *connection) connectTimeout() time.Duration {
	if c.startServer {
		return c.timeout + pob.InstallTimeout
	}
	return c.timeout
}

func (r runner) writeConnection(ctx context.Context, client *grpcapi.Client, c *connection, credentials bool) error {
	info, err := client.Info(ctx)
	if err != nil {
		return err
	}
	if !credentials {
		return r.write(info)
	}
	value, err := c.credentials(client)
	if err != nil {
		return err
	}
	return r.write(value)
}

// credentials передаётся только через явно запрошенный приватный канал stdout.
func (c *connection) credentials(client *grpcapi.Client) (any, error) {
	address, token := c.address, c.token
	if c.profile != "" {
		profile, err := discovery.Open(c.baseDir, c.profile)
		if err != nil {
			return nil, err
		}
		record, err := profile.Read()
		if err != nil {
			return nil, err
		}
		if record.InstanceID != client.InstanceID() {
			return nil, fmt.Errorf("экземпляр сервера изменился при подключении")
		}
		address, token = record.Address, record.Token
	}
	return map[string]string{"address": address, "token": token, "lease_id": client.LeaseID(), "instance_id": client.InstanceID(), "protocol_version": buildinfo.ProtocolVersion}, nil
}

func (r runner) serverCommand(ctx context.Context, command string, args []string) error {
	fs := r.flags(command)
	c := connectFlags(fs)
	c.allowMemory = true
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
	var name, requestID, file string
	var revision uint64
	if operation != "get" {
		fs.StringVar(&name, "name", "", "название сборки")
		fs.StringVar(&requestID, "request-id", "", "идентификатор команды для безопасного повтора")
	}
	switch operation {
	case "create":
		fs.StringVar(&file, "file", "", "начальный XML-файл либо - для stdin")
	case "rename":
		fs.Uint64Var(&revision, "revision", 0, "ожидаемая ревизия")
	}
	if err := parse(fs, args[1:]); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	if operation == "create" {
		cmd := application.CreateSession{SessionID: *id, BuildName: name, RequestID: requestID}
		return r.createSession(ctx, c, cmd, file)
	}
	client, err := c.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	var value domain.Snapshot
	switch operation {
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

func (r runner) createSession(ctx context.Context, c *connection, cmd application.CreateSession, file string) error {
	if file != "" {
		data, err := r.readInput(ctx, file)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return &UsageError{Err: fmt.Errorf("нужен XML сборки")}
		}
		cmd.XML = string(data)
	}
	client, err := c.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	value, err := client.Create(ctx, cmd)
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
	if *standalone {
		return r.standaloneMCP(ctx, c, *id, *name)
	}
	if *id != "" {
		return &UsageError{Err: fmt.Errorf("при --attach укажите сессию только в --attach")}
	}
	return r.attachedMCP(ctx, c, *attach)
}

func (r runner) standaloneMCP(ctx context.Context, c *connection, id, name string) error {
	if c.usesServer() {
		return &UsageError{Err: fmt.Errorf("--standalone не принимает параметры подключения")}
	}
	calculator, closeCalculator, err := openCalculator(ctx)
	if err != nil {
		return err
	}
	defer closeCalculator()
	service := application.NewServiceWithCalculator(calculator)
	if _, err := service.Create(ctx, application.CreateSession{SessionID: id, BuildName: name, RequestID: "standalone-init"}); err != nil {
		return err
	}
	return mcpapi.New(service, id, c.timeout).Run(ctx, &mcp.IOTransport{Reader: r.in, Writer: writer{r.out}})
}

func (r runner) attachedMCP(ctx context.Context, c *connection, id string) error {
	connectCtx, cancel := context.WithTimeout(ctx, c.connectTimeout())
	defer cancel()
	client, err := c.dial(connectCtx)
	if err != nil {
		return err
	}
	defer client.Close()
	if _, err := client.Snapshot(connectCtx, id); err != nil {
		return err
	}
	return mcpapi.New(client, id, c.timeout).Run(ctx, &mcp.IOTransport{Reader: r.in, Writer: writer{r.out}})
}
