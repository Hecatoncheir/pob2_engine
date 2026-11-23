# Приёмка pob2_engine v0.0.12

Выпуск собран из `83b47dab660018c17237ead33a6392f99492dd56`.
Скачанные бинарники повторно запущены на Windows x64, Linux x64 и macOS arm64.
Контрольные суммы всех трёх локальных файлов совпали с результатами раннеров.
[Выпуск](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.12) ·
[Релизный прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45152145898) · [SHA256SUMS](SHA256SUMS).

CLI/gRPC и MCP совпали по пути, режимам узлов и расходу очков. Во втором
комплекте DPS вырос с 19.482916666667 до 24.61. Снятие зависимого пути
восстановило исходный XML.

Dart API PoB II закреплён на коммите кандидата
`b6b886a81fe642d353f583117ac87bd8764441e3`. Приёмка на коммите
`920331c6992560fad4f7bae3b180238717359d1a` проверила смену режима,
повторный предпросмотр после смены оружия через MCP, назначение и снятие ветви.
[Проверки кандидата](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45142216198) ·
[Приёмка PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45143622338).

| Платформа | SHA-256 | Скачанный движок | Общая сборка |
| --- | --- | --- | --- |
| Windows x64 | `7d989808b5750a52e5267d5849f76a6f070ac06d6faa6c968987e0871cce1da2` | [JSON](published_smoke_windows_amd64.json) | [PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45143622338/job/137587413313) |
| Linux x64 | `1dfa7b5d06997e710e8ba74409f7f05f2a60a89abf7eb52c6af41962322fc62e` | [JSON](published_smoke_linux_amd64.json) | [PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45143622338/job/137587413233) |
| macOS arm64 | `2c604fdec89278c3ded3ed171fabee8279abf7092510515c3bba645e6aaf7784` | [JSON](published_smoke_darwin_arm64.json) | [PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45143622338/job/137587412977) |
