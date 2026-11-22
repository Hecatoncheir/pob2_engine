# Приёмка v0.0.6-rc.2

Версия опубликована 22.11.2026. Тег указывает на `b92e9505c21a1f05a905e2741aca3f423482b2a7`.
Все 15 заданий [GitHub Actions](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45043116094) прошли.
[Страница выпуска](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.6-rc.2).

| Платформа | SHA-256 опубликованного бинарника | До публикации | После скачивания |
| --- | --- | --- | --- |
| Windows x64 | `d6c6c34d2c50fa9b5a62d79bdf67e5d5a031bed1dc0f0191b97dd3d943e0ede3` | [отчёт](smoke_windows_amd64.json) | [отчёт](published_smoke_windows_amd64.json) |
| Linux x64 | `b0f948d40d7f32eda6f79a24cdae4e58bf74d3ab60f8c22fd5e1e168fe43c4e6` | [отчёт](smoke_linux_amd64.json) | [отчёт](published_smoke_linux_amd64.json) |
| macOS arm64 | `d64ea6b8192ff24927ef206489dc312ca166e7ebed31d40defda1ba74b803ea7` | [отчёт](smoke_darwin_arm64.json) | [отчёт](published_smoke_darwin_arm64.json) |

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
