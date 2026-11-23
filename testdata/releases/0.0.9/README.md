# Приёмка pob2_engine v0.0.9

[Выпуск](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.9) собран из `2dcf46ef05a902ae9c7cfde944bdb31aa420ac1d`.
[Релизный прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45098489367) проверил запуск скачанных бинарников
с PoB v0.23.1 на каждой платформе. Файлы также сверены с
[SHA256SUMS](SHA256SUMS) и хешами в отчётах CI.

Через CLI и MCP получен одинаковый предпросмотр экипировки. Хеш входа
из сравнения совпал с хешем применённой правки. Процессный тест CLI, MCP
и gRPC также проверил точный повтор правки экипировки после перезапуска.

В [сквозной проверке PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45097834657) коммит `fcdc9d3389910b50c2dbfcd801dded73f7d665b0`
использовал опубликованный кандидат `v0.0.9-rc.2`, закреплённые PoB v0.23.1
и MCP v0.23.0. Проверены примерка, отмена, отказ для неизвестного предмета,
устаревание предпросмотра при правке извне, применение, снятие предмета
и обновление окна после команды MCP. Двоичные файлы финального тега
собраны из кода принятого кандидата.

| Платформа | SHA-256 | Отчёт движка | PoB II |
| --- | --- | --- | --- |
| Windows x64 | `0b2003d25d9f1e6e35f548b3abae95d9c13895e554df156b53e2d4158d8160bd` | [JSON](published_smoke_windows_amd64.json) | [Сквозная проверка](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45097834657/job/137441519187) |
| Linux x64 | `f6b44a9cdbe24605cdeb5995a4356d127aadb0765eee15c7132510c8b522eb6e` | [JSON](published_smoke_linux_amd64.json) | [Сквозная проверка](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45097834657/job/137441518892) |
| macOS arm64 | `15dfcea8a23e28013d2f2d84355f786936f2090933b2b2dd9481280d0ba96ede` | [JSON](published_smoke_darwin_arm64.json) | [Сквозная проверка](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45097834657/job/137441519134) |
