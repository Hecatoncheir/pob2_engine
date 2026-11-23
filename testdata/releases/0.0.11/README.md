# Приёмка pob2_engine v0.0.11

Выпуск собран из `cf7b2ebe834a85c48b55eb3d6083a7318f6a9947`.
Релизный прогон скачал бинарники, сверил хеши и повторил проверки
на Windows x64, Linux x64 и macOS arm64. Все три SHA-256 также сверены
локально. Скачанный Windows-файл
локально прошёл назначение и смену атрибута.
[Выпуск](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.11) ·
[Релизный прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45128833686) · [SHA256SUMS](SHA256SUMS).

CLI/gRPC и MCP показали одинаковый путь. Назначение двух узлов дало
10 ловкости. Смена атрибута узла `13397` дала −5 ловкости и +5 интеллекта.
Состав и порядок узлов сохранились. Хеш применённой правки совпал
с хешем из предпросмотра.

Dart API в ветке PoB II `1.x.x` закреплён на коммите кандидата
`abb3ca7b39e60e978dff03f7837b4500d7fc4a75`. Нативная приёмка PoB II
на коммите `5b3f2119e52aa9761ac3d29302cebf4904527002` проверила назначение, смену,
отмену и блокировку устаревшего предпросмотра после правки через MCP.
[Кандидат](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.11-rc.1) ·
[Проверки кандидата](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45119458607) ·
[Приёмка PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45120323703).

| Платформа | SHA-256 | Отчёт движка | PoB II |
| --- | --- | --- | --- |
| Windows x64 | `b8168e6decc0953d0ade6ef72e0f3c6701aeb243c9748134a4b185ba57dd6ed3` | [JSON](published_smoke_windows_amd64.json) | [Сквозной сценарий](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45120323703/job/137512364277) |
| Linux x64 | `b8389f66e19aa4ce771a5b2923591553a9afa19e458a79aebab386983aa13c55` | [JSON](published_smoke_linux_amd64.json) | [Сквозной сценарий](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45120323703/job/137512364287) |
| macOS arm64 | `fb1531daeacbcb5e843af26ea77da9b88fc0788a722871f48fac72079f650482` | [JSON](published_smoke_darwin_arm64.json) | [Сквозной сценарий](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45120323703/job/137512364386) |
