# Приёмка pob2_engine v0.0.8

[Стабильный выпуск](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.8) собран из `bb880abfa6d9bc302331ad5fd0435fc4530c289b`.
Все 15 заданий [релизного прогона](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45071224985) прошли, включая проверку
скачанных бинарников на Windows x64, Linux x64 и macOS arm64. Задания macOS
до и после публикации потребовали повтора из-за HTTP 403 от GitHub при тесте
отката PoB.

Скачанные файлы отдельно сверены с [SHA256SUMS](SHA256SUMS) и отчётами CI.
Для Ice Nova проверены предпросмотр и добавление Concentrated Area, совпадение
хеша входа, итоговый DPS 10,08 и точный повтор правки после перезапуска.
Проверки используют PoB v0.23.1.

| Платформа | SHA-256 | Нативный отчёт |
| --- | --- | --- |
| Windows x64 | `0c08f3d0181b5c42124b02bdebb57ad44a2884d3a43b5c6440bedccda8a258c0` | [JSON](published_smoke_windows_amd64.json) |
| Linux x64 | `5dc87781f8ce9eebeba612ed174272c98aeac88a5cd1744b538cbe3698c10f5e` | [JSON](published_smoke_linux_amd64.json) |
| macOS arm64 | `ecbb9013695958cc9c59f335cbbf13260f04957c7ecc6b109bd48f1704dcdb6c` | [JSON](published_smoke_darwin_arm64.json) |

В [сквозном прогоне PoB II](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45068983447)
сценарий общей сборки прошёл на трёх ОС с коммитом `b8270f6` и Dart API
кандидата. Результаты этого этапа и диагностика отдельного теста обычного
Lua-редактора описаны в [приёмке кандидата](../0.0.8-rc.1/README.md).
