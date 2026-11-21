# Доказательства приёмки v0.0.2-rc.1

[Кандидат](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.2-rc.1),
commit `b2322bd55165092b5d956d87299a1e3ccaa813ca`.
[Полный workflow](https://github.com/Hecatoncheir/pob2_engine/actions/runs/44870675650)
завершился успешно, включая шесть нативных проверок.

| Платформа | До публикации | Скачанный архив |
| --- | --- | --- |
| Windows amd64 | [smoke](smoke_windows_amd64.json) | [published](published_smoke_windows_amd64.json) |
| Linux amd64 | [smoke](smoke_linux_amd64.json) | [published](published_smoke_linux_amd64.json) |
| macOS arm64 | [smoke](smoke_darwin_arm64.json) | [published](published_smoke_darwin_arm64.json) |

[SHA256SUMS](SHA256SUMS) и [COMMIT.txt](COMMIT.txt) перенесены из опубликованных
вложений, три первых отчёта — тоже из выпуска, три повторных — из артефактов
`published-smoke-*` того же workflow. Их байты сохранены без изменения.
[release.json](release.json) — снимок метаданных GitHub API после скачивания.
[verification.json](verification.json) фиксирует суммы архивов, соответствие
платформ, привязки отчётов и SHA-256 файлов доказательств.

Скачаны все 11 вложений: проверены все десять строк SHA256SUMS и все 11 хешей
API, затем каждый файл внутри трёх платформенных ZIP и неизменность эталона
`0.0.1`. Бинарные ZIP хранятся на странице кандидата; в Git закреплены их суммы
и отчёты. [Сценарий и границы](../../../docs/ACCEPTANCE-0.0.2.md).
