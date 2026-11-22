# Приёмка v0.0.6

Версия опубликована 22.11.2026. Тег указывает на `c1971227794f1f1b828d12eaa7254afa40fa79c1`.
Все 15 заданий [GitHub Actions](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45046608871) прошли.
[Страница выпуска](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.6).

| Платформа | SHA-256 опубликованного бинарника | До публикации | После скачивания |
| --- | --- | --- | --- |
| Windows x64 | `fd103b1509dc32ab8b9bf897395850211cefca0639a4aef57f84727f948159a1` | [отчёт](smoke_windows_amd64.json) | [отчёт](published_smoke_windows_amd64.json) |
| Linux x64 | `cd40a673be778ebd8e66da4960e3f37513ffb9c3128be18c7acbb60b3027357b` | [отчёт](smoke_linux_amd64.json) | [отчёт](published_smoke_linux_amd64.json) |
| macOS arm64 | `27e6c9012d8d88299af1e33ab3dd6fdd09c220fe0ad33517fb3d2498130e159d` | [отчёт](smoke_darwin_arm64.json) | [отчёт](published_smoke_darwin_arm64.json) |

Три бинарника и `SHA256SUMS` скачаны. Размеры и суммы сверены с манифестом
и метаданными GitHub. [Проверка вложений и отчётов](verification.json).

На Windows x64, Linux x64 и macOS arm64 проверены чистая установка,
официальный переход PoB v0.23.0 → v0.23.1, отказ при недоступном обновлении
и офлайн-откат с неисправным запуском новой установки. После перезапуска
сессия восстановилась из базы v0.0.5. Прежний и оба новых ответа с
`request_id` доступны после смены PoB. В отчётах зафиксированы версии,
ревизии, данные каталогов и происхождение результатов.

[Эталоны обеих версий PoB](../../reference/0.0.6/README.md) ·
[Повторные замеры Windows](../../performance/0.0.6-windows/README.md).
