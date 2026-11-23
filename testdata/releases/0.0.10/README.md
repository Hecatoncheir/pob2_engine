# Приёмка pob2_engine v0.0.10

[Выпуск](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.10) собран из `9e2b81e61693385eac8571f8a645b1bbca987fc6`.
[Релизный прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45106535579) скачал и запустил
бинарники на Windows, Linux и macOS. Хеши повторно сверены с
[SHA256SUMS](SHA256SUMS) и отчётами нативных раннеров.

В загруженных бинарниках CLI/gRPC и MCP показали одинаковый путь пассивов.
Применение совпало с предпросмотром. Назначение дало +5 силы и увеличило
DPS Rolling Slam. Снятие ветви восстановило хеш исходного XML.

[Приёмка кандидата](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45103584339) проверила
`v0.0.10-rc.1` из `585dba98729c83fb7b597b7d2191e9a09882066d`. В [приёмке PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45104449521)
ветка `1.x.x` на коммите `5eaba9595fb0be9deadd457242a02265b96cbe8a` скачала этот кандидат
и проверила сквозной сценарий на трёх ОС. Dart API закреплён на коммите
кандидата. Сценарий проверил назначение, снятие, получение новой ревизии
и блокировку устаревшего предпросмотра после команды MCP.

| Платформа | SHA-256 | Отчёт движка | PoB II |
| --- | --- | --- | --- |
| Windows x64 | `7b854022460f2d611a5a8b53509bb1c8e17ce56a704dacb939120a44d842c42b` | [JSON](published_smoke_windows_amd64.json) | [Сквозной сценарий](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45104449521/job/137462404888) |
| Linux x64 | `f773c9d17d430123cbe78ee4e9b98356bfb11f6f054c7d341e45339fb4e7b747` | [JSON](published_smoke_linux_amd64.json) | [Сквозной сценарий](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45104449521/job/137462405019) |
| macOS arm64 | `70ea56f85f5dc28201665db340e24c1769471cb8e73c20354fb700deaa48af5f` | [JSON](published_smoke_darwin_arm64.json) | [Сквозной сценарий](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45104449521/job/137462404836) |
