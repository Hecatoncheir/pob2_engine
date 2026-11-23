# Приёмка pob2_engine v0.0.9-rc.2

[Выпуск](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.9-rc.2) собран из `f078fd9d2e6e952ba2ca1986475da24183aedc94`.
[Релизный прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45096433370) проверил запуск скачанных бинарников
с PoB v0.23.1 на каждой платформе. Файлы также сверены с
[SHA256SUMS](SHA256SUMS) и хешами в отчётах CI.

Через CLI и MCP получен одинаковый предпросмотр экипировки. Хеш входа
из сравнения совпал с хешем применённой правки. Процессный тест CLI, MCP
и gRPC также проверил точный повтор правки экипировки после перезапуска.

В [сквозной проверке PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45097834657) коммит `fcdc9d3389910b50c2dbfcd801dded73f7d665b0`
использовал опубликованный кандидат `v0.0.9-rc.2`, закреплённые PoB v0.23.1
и MCP v0.23.0. Проверены примерка, отмена, отказ для неизвестного предмета,
устаревание предпросмотра при правке извне, применение, снятие предмета
и обновление окна после команды MCP.

| Платформа | SHA-256 | Отчёт движка | PoB II |
| --- | --- | --- | --- |
| Windows x64 | `0e6efa4b9752512768ad7048630b7cfd4a8f7549e3b35a435186f9726c98c563` | [JSON](published_smoke_windows_amd64.json) | [Сквозная проверка](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45097834657/job/137441519187) |
| Linux x64 | `e63f88cf193a9ef26c63a3ca8a71cb8b63d8aff654a17733efa16040c89acb8a` | [JSON](published_smoke_linux_amd64.json) | [Сквозная проверка](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45097834657/job/137441518892) |
| macOS arm64 | `d5a97c2616fa3b897aea81ee3b0a38ce2aceb6ab8d0ec560187bca2b2e4a4ec0` | [JSON](published_smoke_darwin_arm64.json) | [Сквозная проверка](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45097834657/job/137441519134) |
