# Приёмка pob2_engine v0.0.8-rc.1

[Кандидат](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.8-rc.1)
собран из `d9c7a6056fd9dae58c9e2d3e9b34a4bc13c66aa8`.
Все 15 заданий [релизного прогона](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45067921866)
прошли. Проверка macOS перед публикацией прошла после повтора из-за HTTP 403
от GitHub при загрузке предыдущей версии PoB для теста отката.

Нативные проверки повторили расчёт с PoB v0.23.1 на каждой платформе.
Для Ice Nova предпросмотр Concentrated Area совпал с расчётом принятой правки.
Проверены хеш входа, диагностика несовместимой поддержки и точный повтор
после перезапуска. SHA-256 всех скачанных файлов совпали с [SHA256SUMS](SHA256SUMS)
и значениями из отчётов.

| Платформа | Отчёт о скачанном файле |
| --- | --- |
| Windows x64 | [JSON](published_smoke_windows_amd64.json) |
| Linux x64 | [JSON](published_smoke_linux_amd64.json) |
| macOS arm64 | [JSON](published_smoke_darwin_arm64.json) |

Коммиты и хеши зафиксированы в [сводке](verification.json).
[Сквозная проверка PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45068983447)
прошла на Windows x64, Linux x64 и macOS arm64 с коммитом
`b8270f6c67274e6c3d2d403cdb03381dd96e6e4f` и Dart API кандидата.
Отдельный сценарий обычного Lua-редактора в полном Linux-прогоне исчерпал
время ожидания загрузки дерева PoB. Он запускал редактор через `--mcp-editor`
без подключения к движку. [Журнал Linux](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45068983447/job/137351450555).
