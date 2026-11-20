# Работа с полной сборкой

Выпуск `0.0.1` использует [объявленное покрытие Fireball](COVERAGE-0.0.1.md).
Расчёт выполняет Go; Python, Lua, Flutter и сеть при обычном использовании
движка не требуются. Ниже `pob2-engine` — путь к собранному бинарнику;
на Windows используйте `.\pob2-engine.exe`.

## Одноразовый расчёт

```shell
pob2-engine calculate --file build.xml
```

`--file -` читает UTF-8 XML из stdin. Результат JSON направляется в stdout,
диагностика — в stderr. Для такого расчёта сервер не запускается.
Коды выхода: `0` — полный результат, `1` — ошибка выполнения,
`2` — неверные аргументы или XML одноразового расчёта,
`3` — неподдержанный расчёт, `4` — частичный результат.
При `3` и `4` JSON также выводится: проверяйте `status`, `stats`, `issues`.
Ошибка команды сервера, включая конфликт ревизии, возвращает `1`.

## Сохраняемая сессия

```shell
pob2-engine session create --profile main --start-server --session build-1 --name Fireball --request-id create-1
pob2-engine build import --profile main --session build-1 --revision 1 --request-id import-1 --file build.xml
pob2-engine calculate --profile main --session build-1 --revision 2 --request-id calc-1
pob2-engine session get --profile main --session build-1
pob2-engine build export --profile main --session build-1 --output saved.xml
```

Снимок содержит вложенный `build`: XML, нормализованную модель, ревизию входа,
SHA-256 и расчёт. `export --output` создаёт новый файл UTF-8 без перезаписи;
`--raw` выводит XML в stdout. Для новых правок используйте текущую `revision`. Повторяйте
прежний `request_id` только с прежними аргументами. Повтор старой команды
возвращает её исходный снимок; текущую сессию он не откатывает.

## Правки и сравнение

Файл `edit.json` содержит ровно одну правку. Например, увеличение урона:

```json
{"config":{"name":"customMods","value":{"type":"string","value":"100% increased Spell Damage"}}}
```

```shell
pob2-engine build compare --profile main --session build-1 --revision 3 --edit --file edit.json
pob2-engine build edit --profile main --session build-1 --revision 3 --request-id edit-1 --file edit.json
pob2-engine calculate --profile main --session build-1 --revision 4 --request-id calc-2
```

Сравнение не меняет сессию. Без `--edit` сравнивается XML из `--file`.
Любая правка входа очищает текущий расчёт. Примеры других правок:

```json
{"level":{"level":20}}
{"passives":{"nodes":[4739,18845]}}
{"equipment":{"slot":"Ring 1","raw":"Rarity: Normal\nLazuli Ring\nImplicits: 1\n+25 to maximum Mana"}}
{"gems":{"group_id":1,"label":"Fireball","enabled":true,"delete":false,"gems":[{"id":"FireballPlayer","level":1,"quality":0,"enabled":true},{"id":"SupportRapidCastingPlayer","level":1,"quality":0,"enabled":true}]}}
{"selection":{"weapon_set":2}}
```

Каждая строка примера — отдельный JSON-документ. Пустой `equipment.raw`
снимает предмет, `config.value: null` удаляет условие, `gems.delete: true`
удаляет группу. Иные ограничения и гарантии — в [BUILD-COMMANDS.md](BUILD-COMMANDS.md).

## Резервные копии

Сначала остановите сервер, дождитесь его завершения:

```shell
pob2-engine shutdown --profile main
pob2-engine storage backup --profile main --file backup.db
pob2-engine storage verify --file backup.db
pob2-engine storage restore --profile recovered --file backup.db
```

При использовании другого базового каталога повторяйте `--data-dir DIR`.
Восстановление создаёт только новый профиль; существующая база не заменяется.
Подробности миграции и проверки — в [SESSION-STORAGE.md](SESSION-STORAGE.md).

## MCP и gRPC

MCP подключается к существующей сессии:
`pob2-engine mcp --profile main --attach build-1`. Доступны `get_snapshot`,
`rename_build`, `import_build`, `edit_build`, `calculate_build`,
`compare_build`, `export_build`. Аргументы правки совпадают с JSON выше;
внешние поля — `expected_revision`, `request_id`. Импорт получает `xml`.
`compare_build` принимает `xml` либо `edit` и только `expected_revision`.

gRPC предоставляет типизированные модели и те же операции в
[session.proto](../api/pob2/v1/session.proto) и
[build.proto](../api/pob2/v1/build.proto). `CalculateBuild` передаёт этапы,
затем итоговый снимок; последняя версия прогресса может заменять предыдущую.
Для потока нужна аренда клиента. Правки и результат доступны через
[WatchSession](SESSION-EVENTS.md) после commit. Flutter здесь не используется.

## Новые умения в разработке 0.0.2

Исходники разработки поддерживают также Frostbolt и Spark. Это дополнение
к опубликованному `0.0.1`, новый выпуск пока не создан. Примеры из корня:

```shell
go run ./cmd/pob2-engine calculate --file testdata/reference/0.0.2/cases/frostbolt-minimal.xml
go run ./cmd/pob2-engine calculate --file testdata/reference/0.0.2/cases/spark-minimal.xml
```

Ожидаемый `TotalDPS`: `9.1` и `8.5642857142857`. Импорт, правки, сохранение
и сравнение используют прежние команды; показатели зависят от выбранной
стихии и режима. [Покрытие 0.0.2](COVERAGE-0.0.2.md) отделяет готовое от планов.
