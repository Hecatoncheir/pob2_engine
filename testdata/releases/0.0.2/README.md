# Приёмка v0.0.2

[Выпуск v0.0.2](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.2) опубликован 21 ноября 2026 года. Тег указывает на `c99de760578e7f1e13c51fd60d6a57c83578fd2f`. В [запуске GitHub Actions](https://github.com/Hecatoncheir/pob2_engine/actions/runs/44872559781) все 15 заданий завершились успешно.

В каждом задании повторной приёмки раннер скачал архив и `SHA256SUMS` из выпуска, сверил контрольную сумму и запустил сквозной сценарий на целевой системе.

| Платформа | SHA-256 опубликованного архива | Повторная приёмка |
| --- | --- | --- |
| Windows x64 | `ea132e7362fd3ccd0208a7a716ec63add722a54ae57d6a441bcfacc2888b41d5` | [Журнал задания](https://github.com/Hecatoncheir/pob2_engine/actions/runs/44872559781/job/136751031913) |
| Linux x64 | `b30dfa12e5f469461eb841e10e9c4929c0b0f8acc5341ee7f25d48519e1b9eec` | [Журнал задания](https://github.com/Hecatoncheir/pob2_engine/actions/runs/44872559781/job/136751031908) |
| macOS arm64 | `df84db381188467a35f34276497211df3aeda56d9b87643c7a0681fad43b4f08` | [Журнал задания](https://github.com/Hecatoncheir/pob2_engine/actions/runs/44872559781/job/136751031950) |

Сценарий проверил Fireball, Frostbolt и Spark через CLI, MCP и gRPC, сравнение и экспорт, восстановление сессий из резервной копии и совместимость с сохранением `v0.0.1`. Отчёты `published-smoke-*` доступны в артефактах запуска GitHub Actions.
