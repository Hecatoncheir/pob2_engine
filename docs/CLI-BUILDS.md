# Работа с полной сборкой

Расчёты по умолчанию выполняет оригинальный PoB на LuaJIT; движок устанавливает
его автоматически. В 0.0.4 доступны [полный каталог и правки](CATALOG.md),
[примеры новой способности](EXAMPLES-0.0.4.md). Прежний режим Go сохраняет
[покрытие трёх умений](COVERAGE-0.0.2.md) и [свои примеры](EXAMPLES-0.0.2.md).
Python, Flutter и доступ в интернет после установки PoB при обычном использовании
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

Чтобы открыть общую сборку сразу с XML, выполните:

```shell
pob2-engine session create --profile main --start-server --session build-1 --name Fireball --request-id create-1 --file build.xml
```

Первая ревизия уже содержит документ. `--file -` читает XML из stdin.
Ошибочный XML и сбой записи не создают сессию. Точный повтор возвращает
первый снимок, даже если сборка уже изменена.

Существующие команды создания пустой сессии и последующего импорта:

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
{"passives":{"nodes":[54447,4739,18845]}}
{"equipment":{"slot":"Ring 1","raw":"Rarity: Normal\nLazuli Ring\nImplicits: 1\n+25 to maximum Mana"}}
{"gems":{"group_id":1,"label":"Fireball","enabled":true,"delete":false,"gems":[{"id":"FireballPlayer","level":1,"quality":0,"enabled":true},{"id":"SupportRapidCastingPlayer","level":1,"quality":0,"enabled":true}]}}
{"add_support":{"support_id":"SupportRapidCastingPlayer","level":1,"quality":0}}
{"selection":{"weapon_set":2}}
```

Каждая строка примера — отдельный JSON-документ. Пустой `equipment.raw`
снимает предмет, `config.value: null` удаляет условие, `gems.delete: true`
удаляет группу. Иные ограничения и гарантии — в [BUILD-COMMANDS.md](BUILD-COMMANDS.md).
Пример пассивов относится к Sorceress уровня не ниже 3 и включает старт класса.

`add_support` дописывает одну поддержку к выбранной расчётной группе активного
набора. В новой сессии с Ice Nova проверьте Concentrated Area на первой
ревизии:

```shell
pob2-engine session create --profile main --session ice-nova --name "Ice Nova" --request-id ice-create --start-server --file testdata/reference/0.0.4/ice-nova.xml
pob2-engine preview-support --profile main --session ice-nova --revision 1 --support "Concentrated Area" --level 1 --quality 20
```

Если [предпросмотр](CATALOG.md#предпросмотр-полной-группы) подтвердил
совместимость, сохраните `{"add_support":{"support_id":"Concentrated Area","level":1,"quality":20}}`
в `support.json`. В приложении и автоматизации используйте точный
`support.id` из ответа. Укажите ту же ревизию и параметры при записи,
затем рассчитайте изменённую сборку:

```shell
pob2-engine build edit --profile main --session ice-nova --revision 1 --request-id ice-support --file support.json
pob2-engine calculate --profile main --session ice-nova --revision 2 --request-id ice-calculation
```

В MCP передайте тот же объект в поле `edit` инструмента `edit_build`.

### Примерка предмета

Сохраните правку `equipment` в `equipment.json`. Пример имени слота PoB —
`Body Armour` или `Ring 1`. Вставьте текст из игры или PoB. Чтобы снять предмет,
передайте пустую строку в `raw`. Сначала сравните, затем примените к той же
ревизии, если вариант подходит:

```shell
pob2-engine build compare --profile main --session build-1 --revision 3 --edit --file equipment.json
pob2-engine build edit --profile main --session build-1 --revision 3 --request-id equip-1 --file equipment.json
```

Неизвестная база или несовместимый слот приводят к отказу. Если PoB вернул
числа с предупреждениями, CLI печатает результат и завершает сравнение
с кодом `4`. Через MCP передавайте `edit.equipment` в `compare_build` и
`edit_build`, через gRPC — `EquipmentEdit` в `CompareBuild` и `ApplyBuild`.

### Правка пассива

Получите ID из каталога `passives`. Сохраните, например,
`{"tree_node":{"node_id":13397,"action":"allocate"}}` в `passive.json`.
Для снятия укажите `refund`. Предпросмотр возвращает затронутые узлы
и изменение показателей, а при успешной записи меняется активное дерево.

```shell
pob2-engine build compare --profile main --session build-1 --revision 3 --edit --file passive.json
pob2-engine build edit --profile main --session build-1 --revision 3 --request-id tree-1 --file passive.json
```

В MCP передайте `edit.tree_node` в `compare_build` или `edit_build`.
В gRPC передайте `TreeNodeEdit` в `CompareBuild` или `ApplyBuild`.

Для назначения узла с ловкостью запишите в `passive.json`
`{"tree_node":{"node_id":13397,"action":"allocate","attribute":"dexterity"}}`.
Чтобы заменить ловкость на интеллект у назначенного узла, после
назначения возьмите новую ревизию и передайте
`{"tree_node":{"node_id":13397,"action":"set_attribute","attribute":"intelligence"}}`.
Ход назначения и изменения атрибутов указаны в `passive_change`.

Для назначения в ветвь второго комплекта передайте
`{"tree_node":{"node_id":51184,"action":"allocate","allocation_mode":"weapon_set_2"}}`.
Для снятия узла из этой ветви передайте `action: "refund"` с тем же режимом.
При предупреждениях Lua сравнение выводит JSON и код 4.

Пустое гнездо принимает ту же команду с общим режимом, например
`{"tree_node":{"node_id":26725,"action":"allocate","allocation_mode":"common"}}`.
Для снятия укажите `refund`. При попытке изменить занятое гнездо Lua
возвращает ошибку до записи. Контракт доступен через `build.tree_sockets`.

Для назначенного гнезда используйте `{"jewel":{"node_id":26725,"item_id":2}}`.
Вместо `item_id` можно передать `raw` с текстом нового самоцвета. Укажите
только `node_id`, чтобы снять самоцвет и сохранить его в списке предметов.
Передайте файл правки в `build compare --edit --file` для предпросмотра
или в `build edit --file` для записи. MCP принимает ту же форму в `edit`,
gRPC использует `BuildEdit.jewel` и объявляет `build.tree_jewels`.

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
`compare_build`, `export_build`, `get_catalog`, `preview_support`. Аргументы правки совпадают с JSON выше;
внешние поля — `expected_revision`, `request_id`. Импорт получает `xml`.
`compare_build` принимает `xml` либо `edit` и только `expected_revision`.
У каталога есть `kind/query/offset/limit`; предпросмотр получает
`support_id/expected_revision/level/quality`. [Матрица прежних инструментов](MCP-MATRIX.md)
отделяет ограниченные аналоги от ещё не перенесённых функций.

gRPC предоставляет типизированные модели и те же операции в
[session.proto](../api/pob2/v1/session.proto) и
[build.proto](../api/pob2/v1/build.proto),
[catalog.proto](../api/pob2/v1/catalog.proto). `CalculateBuild` передаёт этапы,
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

Для выбора поддержек доступны `catalog --kind gems --query Support` и
`preview-support --session build-1 --revision 2 --support "Rapid Casting I"`
с обычными флагами подключения. Предпросмотр сравнивает полную группу без записи,
возвращая JSON и те же коды статусов 0/3/4. Каталог содержит также `classes`,
`items`, `passives`; лимит страницы — 100. [Контракт и ограничения](CATALOG.md).
