# Приёмка pob2_engine v0.0.14

Выпуск собран из коммита `2c3ecec7852a9c0ec9325de34c102f3b299ca8a0`.
GitHub Actions сверил скачанные бинарники с SHA-256 и запустил их
на Windows x64, Linux x64 и macOS arm64.
[Выпуск](https://github.com/Hecatoncheir/pob2_engine/releases/tag/v0.0.14) ·
[Релизный прогон](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45285040103).

| Платформа | SHA-256 | Скачанный бинарник |
| --- | --- | --- |
| Windows x64 | `1ae511126b6ebe70fda9a0c3e86156cc60c06ca510e2c898c81bf59c5fd3ac8e` | [Нативная проверка](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45285040103/job/138087548841) |
| Linux x64 | `03483e1d08cc55524e15086502f11dfb81b39a5bd9e13c0df7f8aee249e95d62` | [Нативная проверка](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45285040103/job/138087548764) |
| macOS arm64 | `8fcba44be3ef5b57c3ef64c586b5477367b4d7d6a872ea3a325f05273d1aba14` | [Нативная проверка](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45285040103/job/138087548797) |

Проверка мастерской импортирует XML, сверяет урон Fireball с эталоном
исходного Lua, экспортирует сборку и получает PNG сводки и полного дерева.
Чтение и рисование сохраняют ревизию. Рабочий процесс запускается с пустым
`PATH`.

На Windows Go-клиент прочитал XML сборки 90-го уровня и папку сборок
из окна изолированной переносимой копии оригинального PoB. После снятия
моста SHA-256 файла Launch.lua совпал с исходным. Отдельный профиль
PoB II без Python MCP восстановил прежнюю общую сессию, заметки и
побайтово тот же XML. Go MCP объявил 42 инструмента, вернул PNG окна,
сохранил XML и завершил PoB II с кодом 0.

Dart API закреплён на `708523bab034964087d018ec7fa913f9a06c4ce9`.
[Приёмка RC2](https://github.com/Hecatoncheir/pob2_engine/actions/runs/45282831792) ·
[Приёмка PoB II со стабильным движком](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45294308966).
