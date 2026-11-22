# Приёмка pob2_engine v0.0.7-rc.1

[Кандидат](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.7-rc.1)
собран из `721428348fa21fc4b610fe2fc14ad137f178524a`.
[Релизный прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45058491477)
проверил Windows x64, Linux x64 и macOS arm64 до и после публикации.
Задания macOS повторены после HTTP 403 при проверке обновления PoB через GitHub.

Во всех сценариях бинарник создаёт первую ревизию из XML одной командой.
Повтор команды после перезапуска возвращает исходный снимок, а база
[выпущенного v0.0.6](../../compat/0.0.6/README.md) сохраняет прежний результат
расчёта и принимает новый с тем же PoB v0.23.1.

| Платформа | Сборка CI | Скачанный бинарник |
| --- | --- | --- |
| Windows x64 | [Отчёт](smoke_windows_amd64.json) | [Отчёт](published_smoke_windows_amd64.json) |
| Linux x64 | [Отчёт](smoke_linux_amd64.json) | [Отчёт](published_smoke_linux_amd64.json) |
| macOS arm64 | [Отчёт](smoke_darwin_arm64.json) | [Отчёт](published_smoke_darwin_arm64.json) |

Комплект проверен по [SHA256SUMS](SHA256SUMS). Сводка и ссылки записаны
в [verification.json](verification.json). Сквозная проверка приложения
PoB II выполняется отдельным прогоном в репозитории клиента.
