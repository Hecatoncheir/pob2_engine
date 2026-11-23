# Лицензии и происхождение материалов

Новый код `pob2_engine` распространяется по [MIT](LICENSE), copyright © 2026
Hecatoncheir. Эта лицензия не заменяет условия сторонних материалов.

## Зависимости исполняемого файла

Точный состав Go-модулей CLI для Windows x64, Linux x64 и macOS arm64,
версии, пути уведомлений и SHA-256 находятся в
[манифесте](.licenses/runtime/manifest.json). Исходные тексты `LICENSE`,
`NOTICE`, `AUTHORS` и `PATENTS` сохранены без изменения в `.licenses/runtime/`.
Уведомления стандартной библиотеки и её vendor-пакетов учтены отдельно.

Основные условия: MIT, BSD и Apache-2.0. `modelcontextprotocol/go-sdk` содержит
собственное уведомление о переходе лицензий; применим весь сохранённый
[текст upstream](.licenses/runtime/modules/github.com/modelcontextprotocol/go-sdk/v1.8.0/LICENSE).
Не следует заменять его одной краткой меткой. Полный граф `go.mod` включает
также инструменты и тестовые зависимости; он не равен составу бинарника.

```shell
python tool/notices.py --write
python tool/notices.py
python tool/notices.py --archive .reference/licenses.zip
```

Первая команда обновляет копии из проверенного кэша Go; вторая обнаруживает
новые зависимости, изменённые тексты и устаревшие файлы. Используйте версию Go
из `go.mod`. CI проверяет копии. Релизный бинарник выдаёт лицензии по команде
`doctor --licenses`. При удалении зависимости сначала
проверьте указанные скриптом оставшиеся уведомления.

## Расчётный адаптер и эталон

Файлы `HeadlessWrapper.lua`, `_SimpleGraphic.def.lua`, `headless_bootstrap.lua`,
`tree-image-sizes.json`, `gem_groups.lua`, `granted_supports.lua`, `skill_group.lua`
и части отрисовки `tree_scene.lua` и `worker.lua`
в `internal/pob/bridge/` основаны на
[pob-poe2-mcp, commit 312b337fa4a24f5ef55ebd38c8fd7bd46c7de78a](https://github.com/Hecatoncheir/pob-poe2-mcp/tree/312b337fa4a24f5ef55ebd38c8fd7bd46c7de78a/pob_poe2_mcp)
и распространяются по [MIT](internal/pob/bridge/LICENSE). Оригинальные Lua-код
и данные PoB загружаются из выпуска сообщества. LuaJIT и `lua-utf8` для Linux
и macOS загружаются из [pob-poe2-mcp v0.24.1](https://github.com/Hecatoncheir/pob-poe2-mcp/releases/tag/v0.24.1).

Совместимые сценарии мастерской и Lua-мост оригинального окна используют
файлы [pob-poe2-mcp v0.23.0, commit 1b3f61dc00016fe3bc86d889d6f4992302231d75](https://github.com/Hecatoncheir/pob-poe2-mcp/tree/1b3f61dc00016fe3bc86d889d6f4992302231d75/pob_poe2_mcp).
Исходные Lua-модули, описания инструментов и таблицы ресурсов сохранены
в `internal/poblegacy/assets/` вместе с [лицензией MIT](internal/poblegacy/assets/LICENSE).

Закреплённые эталонные копии PoB и Python/Lua MCP используются в разработке и CI.
Ревизии и контрольные
суммы заданы в [sources.lock.json](testdata/reference/sources.lock.json);
оригинальные лицензии сохранены отдельно:
[PoB](.licenses/reference/pob-LICENSE.md), [MCP](.licenses/reference/mcp-LICENSE.txt).
Объединённый документ PoB содержит также уведомления его сторонних компонентов.

При переносе алгоритма или данных запишите исходный репозиторий, полный commit,
путь, SHA-256, применимую лицензию и характер изменений в манифесте набора.
Сохраняйте исходные copyright и уведомления; адаптированный алгоритм помечайте
как изменённый. До включения в поставку проверьте условия конкретного файла.

Данные и изображения игры учитываются отдельно от кода PoB: наличие MIT у
репозитория не означает передачи прав на любые игровые материалы. Встроенный
[числовой набор](docs/DATASET.md) перечислен в
[манифесте](internal/gamedata/files/manifest.json): закреплённые исходные пути,
SHA-256, copyright Grinding Gear Games и преобразования. Изображения и описания
умений не включены. Ограниченный перенос формул ресурсов, попаданий и округления
PoB отмечен в `internal/calculation/compute.go`; исходные пути, ревизия,
copyright, лицензия и SHA-256 приведены в
[origin.json](internal/calculation/origin.json). Авторские минимальные XML и результаты эталонного расчёта
хранятся в `testdata/reference/`; их происхождение фиксируется рядом.

## Проверка для разработки 0.0.2

Источники остаются закреплёнными: PoB `v0.23.1` / `7d6f530cbdab20389ff8bc6ba97a37ac27f74e41`,
MCP `v0.24.1` / `312b337fa4a24f5ef55ebd38c8fd7bd46c7de78a`.
Числовой набор `pob2-0.0.2-spell-hits-v1` имеет схему 2, дерево `0_5`;
`game_version=null` не объявляет версию игры. Источником расчёта остаётся
явно закреплённый PoB, а не иная ревизия во внутреннем lock-файле MCP.

Проверяемые материалы:

| Материал | Происхождение и условия |
| --- | --- |
| Числа трёх умений, классов, украшений и пассивов | Пути, SHA-256, преобразования и уведомления GGG в [manifest.json](internal/gamedata/files/manifest.json); точное воспроизведение через tool/export_data.py |
| Перенесённые формулы | Источники и изменения в [origin.json](internal/calculation/origin.json), MIT и сохранённая лицензия PoB |
| 334 эталонных случая Lua | [Закреплённые источники и runtime](testdata/reference/README.md), отдельные XML/JSON и числовые допуски |
| Таблица прежних MCP-инструментов | [Манифест закреплённой ревизии и его SHA-256](docs/MCP-MATRIX.md); сопоставляются контракты, исходники прежнего сервера не включаются в runtime |
| Зависимости CLI | 15 модулей, 33 файла уведомлений в сохранённом манифесте; проверка tool/notices.py для трёх целевых платформ |

Новые зависимости, игровые изображения и заимствования при обновлении
документации этапа 7.1 не добавляются. Условия Go-кода, исходных материалов
PoB и игровых данных продолжают учитываться раздельно.

21 ноября 2026 года проверены `tool/notices.py`, `go run ./tool/reference` и
`tool/export_data.py`: комплект уведомлений совпадает с runtime, закреплённые
источники проверены, оба файла набора воспроизведены без изменения байтов.
