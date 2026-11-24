# Приёмка pob2_engine v0.0.13

Выпуск собран из `c07907f27cee82e2ceb388d2ac1ffc1e2df85236`.
Опубликованные бинарники сверены с SHA-256 и запущены на Windows x64,
Linux x64 и macOS arm64.
[Выпуск](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.13) ·
[Релизный прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45267078153).

| Платформа | SHA-256 | Проверка скачанного бинарника |
| --- | --- | --- |
| Windows x64 | `93019459c53f3ae1d5d68bc217d4c93e263935767e0fe7718ecdec3746b39b55` | [Прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45267078153/job/138025333521) |
| Linux x64 | `ad8f2f5a40d734794e74843873ff1d9cc6138362c09b2b5c4567a4e6a220da3f` | [Прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45267078153/job/138025333223) |
| macOS arm64 | `9bf218c676b9591c37fd0bec273b0ff4b1413538c0230343db5ec50b5fe486f2` | [Прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45267078153/job/138025333294) |

Приёмка проверила правки классов и пассивов, Lua-отчёты и редакторы.
После добавления Rapid Casting I к умению Chaos Bolt от Withered Wand
движок восстановил группу из сохранённого XML.

Dart API PoB II закреплён на `708523bab034964087d018ec7fa913f9a06c4ce9`.
[Приёмка общей мастерской с RC3](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45263016488)
прошла на трёх ОС. Следующая приёмка PoB II 1.x проверяет запуск
без Python MCP с движком v0.0.14.
