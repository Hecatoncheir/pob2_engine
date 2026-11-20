# Итог выпуска 0.0.1

21 ноября 2026 года по Москве опубликован
[v0.0.1](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.1),
commit `d5bb6a2a48fcb6f9b888eb334781fdd6948dc589`.
[Полный релизный workflow](https://github.com/Hecatoncheir/pob2_engine/actions/runs/44835654049)
успешен: Linux CI, Go/Lua, контракт, три кросссборки, запуск архивов,
публикация и повторный запуск файлов, скачанных с GitHub Release.

## Проверенные файлы

Скачаны все 11 файлов выпуска: десять совпали с `SHA256SUMS`, сам список
сумм сохранён отдельно. В каждом платформенном ZIP проверены все записи
`PACKAGE.json`, версия, commit и SHA-256 содержимого. Проверены также архивы
исходников ZIP/tar.gz и комплект лицензий.

| Платформа | Архитектура | Повторный запуск скачанного архива |
| --- | --- | --- |
| Windows Server 2025 | amd64 | [Успешен](../testdata/releases/0.0.1/published_smoke_windows_amd64.json) |
| Linux, ядро 7.0.0-1020-azure | amd64 | [Успешен](../testdata/releases/0.0.1/published_smoke_linux_amd64.json) |
| macOS 15.8.2 | arm64 | [Успешен](../testdata/releases/0.0.1/published_smoke_darwin_arm64.json) |
| Windows 11, локальный запуск | amd64 | [Успешен](../testdata/releases/0.0.1/published_smoke_windows11_local.json) |

Проверены расчёт с пустым PATH, общая сессия CLI/MCP через gRPC, изменение
урона, сравнение без правки, экспорт XML, Backup/Verify/Restore и точный
повтор сохранённого запроса. Движок работает без Python/Lua/PoB; Python
использован только как проверяющий инструмент. Исходный пример даёт
`TotalDPS = 8.9166666666667`, после `100% increased Spell Damage` —
`17.8333333333333`, в пределах относительного допуска `1e-10` и абсолютного `1e-9`.

Исходные [SHA256SUMS](../testdata/releases/0.0.1/SHA256SUMS) и
[COMMIT.txt](../testdata/releases/0.0.1/COMMIT.txt) сохранены в репозитории.
Установка и команды: [RELEASE.md](RELEASE.md), [CLI-BUILDS.md](CLI-BUILDS.md).
Матрица 77 вариантов и проверки отказов: [приёмка](ACCEPTANCE-0.0.1.md).
[Производительность](../testdata/performance/README.md) измерена отдельно
для библиотеки; ускорение относительно Lua не заявляется.

## Ограничения

Поддерживается [объявленный набор Fireball Projectile](COVERAGE-0.0.1.md):
Sorceress без возвышения, Fireball 1–20 качества 0, две поддержки,
выбранные украшения, свойства и начальная ветка пассивов. Это самостоятельный
законченный сценарий, а не полный паритет с PoB. Неизвестные активные механики
дают `unsupported` без чисел; предупреждения доступны в `issues`.
Flutter не изменялся и не участвовал в приёмке. Подпись исполняемых файлов
и notarization macOS не выполнялись. Совпадение с PoB не означает проверку в игре.
