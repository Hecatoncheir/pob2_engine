# Сервис сессий

Реализованы Go CLI, локальный gRPC API и MCP stdio на официальном [Go SDK](https://github.com/modelcontextprotocol/go-sdk). Создание сессии, чтение снимка и переименование используют общий сервис `internal/application`. Сессии профиля [сохраняются на диск](SESSION-STORAGE.md) и восстанавливаются после перезапуска. Работают [обнаружение, автозапуск и аренды клиентов](SERVER-LIFECYCLE.md). Игровых расчётов, импорта XML, подписок и подключения Flutter пока нет.

## Запуск и подключение

Нужен Go 1.27.4. Из корня репозитория:

```powershell
go build ./cmd/pob2-engine
.\pob2-engine.exe --version
.\pob2-engine.exe serve
```

На Linux и macOS запускайте `./pob2-engine`. Сервер выбирает свободный порт на `127.0.0.1` и выводит одну строку JSON с `address`, случайным `token`, `protocol_version` и `mode: persistent`. Этот вывод содержит секрет подключения: передавайте его только своим клиентам. Журналы и ошибки направляются в stderr.

Без имени `serve` использует сохраняемый профиль `default`. `storage_mode` и `storage_path` в готовности и `info` сообщают способ и путь хранения; `--ephemeral` выбирает временный сервер без профиля.

Во втором терминале подставьте адрес и токен из вывода сервера:

```powershell
$address = '127.0.0.1:ПОРТ'
$env:POB2_ENGINE_TOKEN = 'ТОКЕН'
.\pob2-engine.exe info --address $address
.\pob2-engine.exe session create --address $address --session main --name 'Первая сборка' --request-id create-main-1
.\pob2-engine.exe session get --address $address --session main
.\pob2-engine.exe session rename --address $address --session main --name 'Новое название' --revision 1 --request-id rename-main-1
.\pob2-engine.exe shutdown --address $address
```

Вместо переменной окружения допускается `--token`. Для сервера можно явно задать `--listen 127.0.0.1:50051`. Другие интерфейсы запрещены. Подключение всегда требует адрес и токен; таймаут команд задаётся через `--timeout 5s`. Ошибка подключения или отсутствующая сессия не создаёт независимую сборку. Коды завершения CLI: `0` — успех, `1` — ошибка операции, `2` — ошибка синтаксиса.

Вместо ручных адреса и токена можно указать `--profile main`: данные берутся из защищённого файла подключения. Для явного автозапуска добавьте `--start-server`; длительная команда `connect --profile main --start-server` удерживает automatic-сервер до EOF или Ctrl+C. Режим, блокировки, регистрация и завершение описаны в [отдельном руководстве](SERVER-LIFECYCLE.md).

## Контракт и команды

Контракт находится в [session.proto](../api/pob2/v1/session.proto). `GetServerInfo` сообщает версию программы, версию протокола `pob2.v1`, состояние и поддерживаемые возможности. Go-клиент проверяет протокол перед командами. Dart-клиенту при подключении также нужно вызвать этот метод и передавать metadata `authorization: Bearer ТОКЕН` для каждого RPC.

Манифест `gen/dart/pubspec.yaml` сохраняйте в ASCII: при чтении Git-зависимости
на Windows `pub` может использовать системную кодировку для вывода Git.
Документация и исходники API остаются в UTF-8.

| RPC | Результат |
| --- | --- |
| `CreateSession` | Сессия с явным `session_id`, названием и ревизией `1` |
| `GetSnapshot` | Копия текущего состояния |
| `RenameBuild` | Правка при совпадении `expected_revision`, затем увеличение ревизии |
| `WatchSession` | Начальный снимок или правки после заданной ревизии, затем новые события; см. [руководство](SESSION-EVENTS.md) |
| `Shutdown` | Переход в `draining` и завершение сервера |
| `RegisterClient`, `Heartbeat`, `ReleaseClient` | Регистрация, подтверждение активности и освобождение аренды клиента |

`session_id` и `request_id` содержат 1–128 символов ASCII: буквы, цифры, `.`, `_`, `-`; первый символ — буква или цифра. Название — непустая строка UTF-8 до 512 байт. Создание использует уникальный `request_id` в пределах сохранённого профиля; переименование — в пределах сессии, с запретом повторить идентификатор её создания.

Точный повтор успешной команды возвращает её исходный снимок, даже если сборка уже изменена другими командами. Повтор с другими аргументами отклоняется. На профиль допускается 1024 сессии, на сессию — 4096 успешных переименований. История результатов сохраняется после перезапуска. При ошибке commit команда может остаться неподтверждённой: после восстановления нужно повторить тот же `request_id`, а не создавать новый.

| Ошибка | gRPC status |
| --- | --- |
| Некорректные аргументы | `InvalidArgument` |
| Нет сессии / сессия уже существует | `NotFound` / `AlreadyExists` |
| Конфликт ревизии | `Aborted` |
| Повтор идентификатора с другими аргументами | `FailedPrecondition` |
| Исчерпан лимит | `ResourceExhausted` |
| Неверный токен | `Unauthenticated` |
| Сервер завершает работу | `Unavailable` |
| Ошибка сохранения (`storage_failure`) | `Unavailable` |

## MCP

Для общей сессии клиент запускает:

```powershell
.\pob2-engine.exe mcp --attach main --address $address
```

Токен наследуется через `POB2_ENGINE_TOKEN`. Доступны два инструмента: `get_snapshot` без аргументов и `rename_build`:

```json
{"build_name":"Название из MCP","expected_revision":1,"request_id":"mcp-rename-1"}
```

Ответ содержит структурированный снимок; ошибка инструмента возвращается с `isError: true`. Сессия выбирается при запуске MCP и не меняется аргументами инструментов. Initialize, согласование MCP и `tools/list` выполняет SDK. stdout содержит только сообщения MCP.

Для явно независимой сессии используйте `mcp --standalone --session private --name 'Локальная'`. Закрытие stdin завершает MCP-процесс. Отключение attach-клиента сохраняет работу постоянного сервера.

## Проверки и генерация

```shell
go vet ./...
go test ./...
go test -race ./...
```

Процессные тесты входят в `go test ./...`: они собирают CLI для тестов, запускают сервер и MCP отдельно, проверяют общее состояние, повторы, конфликты и завершение через `shutdown` / EOF. Все тесты CI выполняются только на Linux. Файлы для Windows x64, Linux x64 и macOS arm64 с `CGO_ENABLED=0` собираются только релизным workflow по тегу `v*`. `-race` требует доступного C-компилятора. Отдельный тест проверяет ограничение времени завершения активного RPC. Ctrl+C и SIGTERM используют тот же механизм остановки; принудительное уничтожение процесса его обходит. На завершение даётся до пяти секунд, затем RPC прерываются.

Для генерации нужны Dart 3.14.0 и [Buf](https://github.com/bufbuild/buf) 1.73.0:

```shell
go install github.com/bufbuild/buf/cmd/buf@v1.73.0
buf format --diff --exit-code
buf lint
go run ./tool/generate
go run ./tool/generate -check
dart pub -C gen/dart get --enforce-lockfile
dart analyze gen/dart
```

Добавьте каталог `bin` из `go env GOPATH` в PATH. Версии Go-плагинов закреплены в `buf.gen.yaml`, Dart-плагина и его зависимостей — в `tool/proto/pubspec.yaml` и lock-файле. Генерация локальная, без удалённых плагинов Buf. Проверка `-check` генерирует API во временный каталог и сравнивает содержимое и набор файлов с `gen/go` и `gen/dart/lib`. На Windows используется Dart SDK из Flutter, минуя служебный вывод обёртки. Версии [gRPC Go](https://grpc.io/docs/languages/go/quickstart/) и Protobuf закреплены в `go.mod`; Dart API оформлен отдельным пакетом с lock-файлом. CI проверяет совместимость схемы с `main`, когда там уже есть базовый контракт.
