# Приёмка pob2_engine v0.0.17

Выпуск собран из коммита `125933399ac85cfafe9b7bdf156709a3ab9695ae`. GitHub Actions проверил
CLI, MCP, gRPC, работу с оригинальным Lua, сохранение и восстановление
сессий до публикации и повторил проверку со скачанными бинарниками.
SHA-256 всех трёх бинарников совпали с опубликованным `SHA256SUMS`.
[Выпуск](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.17) · [Релизный прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45527885277).

| Платформа | SHA-256 стабильного бинарника | Проверка скачанного файла | Приёмка PoB II с кандидатом |
| --- | --- | --- | --- |
| Windows x64 | `598546dc2b75635183131e1bc1c586066e4e15f287189b4b9d4dca029a8645c9` | [Бинарник](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45527885277/job/138925995206) | [PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45521635792/job/138894693132) |
| Linux x64 | `42b89c6df98aa50482facdae95ffd48b359beb451b0e4febebd03c69de96046d` | [Бинарник](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45527885277/job/138925995208) | [PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45521635792/job/138894693167) |
| macOS arm64 | `b37561c18eaa3d950967406ff348ffdb49520a9908140f4ca4dad507e823796b` | [Бинарник](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45527885277/job/138925993565) | [PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45521635792/job/138894693067) |

В первой попытке Linux эталонный тест влияния пассивов вернул различие
между отчётами. На том же коммите эталонная проверка прошла при повторе
и релизный прогон завершился успешно. [Первая попытка](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45527885277/attempts/1).

Таймаут MCP при первой проверке скачанного файла macOS не повторился.
Повторная проверка того же опубликованного бинарника завершилась успешно.

## PoB II

Полная приёмка PoB II с комплектным `v0.0.17-rc.2` прошла из коммита
`4f1bbe6389eb220c2bd1c47bbc3e5768e3ec01d6`. На трёх ОС проверены общая XML-сборка, мастерская,
установка и восстановление с Lua, очистка профиля, PNG через MCP
и штатное завершение приложения. [Приёмка PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45521635792).

После проверки стабильного выпуска бинарник `v0.0.17` закреплён в ветке
`1.x.x` PoB II в коммите `c862adfaa5a28fde5b09698530fb14571aa1b28a`. Все 17 файлов сгенерированного Dart
API совпадают с файлами стабильного тега движка.
[Закрепление версии](https://github.com/Hecatoncheir/pob2_ui/commit/c862adfaa5a28fde5b09698530fb14571aa1b28a).

Тег `v0.0.17-rc.1` сохранён. Публикация его файлов остановилась на
проверке контракта Protobuf. [Проверка rc.1](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45510800086) ·
[Выпуск rc.2](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.17-rc.2).
