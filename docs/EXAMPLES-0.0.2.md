# Примеры трёх умений в 0.0.2

Примеры относятся к исходникам разработки. Проверяются объявленные попадания
по одной цели, базовый класс Sorceress, камень уровня 1/качества 0, без предметов
и наград заданий. Это минимальные примеры API, а не рекомендации игрового билда.
[Ограничения](COVERAGE-0.0.2.md) и `issues` остаются частью результата.

## Одноразовые расчёты

Из корня репозитория:

```shell
go run ./cmd/pob2-engine calculate --file testdata/reference/fireball.xml
go run ./cmd/pob2-engine calculate --file testdata/reference/0.0.2/cases/frostbolt-minimal.xml
go run ./cmd/pob2-engine calculate --file testdata/reference/0.0.2/cases/spark-minimal.xml
```

| Умение | `mode` результата | `TotalDPS` |
| --- | --- | --- |
| Fireball | Projectile hits | 8.9166666666667 |
| Frostbolt | Projectile hits | 9.1 |
| Spark | Base hits | 8.5642857142857 |

Ожидаются `status: complete`, непустые `stats`, хеш входа и
`data_id: pob2-0.0.2-spell-hits-v1`. Число снарядов, взрывы и повторные попадания
Spark не умножают эти показатели. CLI возвращает 0/3/4 для complete/unsupported/partial.

## Сохраняемая сессия Frostbolt

Соберите `go build ./cmd/pob2-engine`. Ниже `pob2-engine` означает путь к бинарнику:
`./pob2-engine` на Linux/macOS, `.\pob2-engine.exe` на Windows. Выполняйте команды
из корня, в новом профиле `guide-002`.

Запустите постоянный сервер в отдельном терминале и оставьте его открытым:

```shell
pob2-engine serve --profile guide-002
```

Остальные команды выполняйте во втором терминале:

```shell
pob2-engine session create --profile guide-002 --session frostbolt --name Frostbolt --request-id create-frostbolt
pob2-engine build import --profile guide-002 --session frostbolt --revision 1 --request-id import-frostbolt --file testdata/reference/0.0.2/cases/frostbolt-minimal.xml
pob2-engine catalog --profile guide-002 --kind gems --query RapidCastingSupport --limit 5
pob2-engine preview-support --profile guide-002 --session frostbolt --revision 2 --support "Rapid Casting I"
pob2-engine session get --profile guide-002 --session frostbolt
```

Предпросмотр оставляет ревизию **2** и не сохраняет расчёт. Его `before` имеет
`TotalDPS=9.1`, `after` показывает влияние добавленной поддержки. Исходный XML
не меняется. Установка поддержки выполняется отдельной правкой `gems`, которая
заменяет весь состав группы; перечисляйте все сохраняемые камни.

## MCP в той же сессии

После предыдущих команд запустите MCP из своего клиента:

```shell
pob2-engine mcp --profile guide-002 --attach frostbolt
```

Далее показаны объекты `params` запросов `tools/call`; согласование MCP и
оболочку JSON-RPC выполняет клиент. Выполняйте примеры один раз по порядку.

```json
{"name":"get_snapshot","arguments":{}}
```

Установите явно видимое предположение +100% Spell Damage. Это пользовательское
условие, а не надетый предмет:

```json
{"name":"edit_build","arguments":{"expected_revision":2,"request_id":"damage-frostbolt","edit":{"config":{"name":"customMods","value":{"type":"string","value":"100% increased Spell Damage"}}}}}
```

Ответ содержит ревизию **3**, старый расчёт очищен. Рассчитайте и сохраните:

```json
{"name":"calculate_build","arguments":{"expected_revision":3,"request_id":"calc-frostbolt"}}
```

Ожидаются ревизия **4**, `build.calculation.status=complete`, `TotalDPS=18.2`.
Предпросмотр и сравнение после этого используют ревизию 4:

```json
{"name":"preview_support","arguments":{"expected_revision":4,"support_id":"Rapid Casting I"}}
```

```json
{"name":"compare_build","arguments":{"expected_revision":4,"edit":{"config":{"name":"customMods","value":{"type":"string","value":"200% increased Spell Damage"}}}}}
```

Сравнение возвращает `before.TotalDPS=18.2`, `after.TotalDPS=27.3` внутри
`comparison.*.stats`. Сессия остаётся в ревизии 4 с условием +100%.
Повтор `calc-frostbolt` с теми же аргументами возвращает первоначальный ответ;
смена входа требует нового `request_id` и актуальной ревизии из `get_snapshot`.

CLI читает тот же сохранённый результат и может экспортировать XML:

```shell
pob2-engine session get --profile guide-002 --session frostbolt
pob2-engine build export --profile guide-002 --session frostbolt --raw
pob2-engine shutdown --profile guide-002
```

Профиль сохраняется после завершения. При следующем подключении добавьте
`--start-server`, если сервер уже остановился. Для Fireball и Spark используйте
соответствующий XML из первого раздела и отдельный `session_id`/ID создания.
Полный перечень инструментов и различий прежнего MCP — в [матрице](MCP-MATRIX.md).
