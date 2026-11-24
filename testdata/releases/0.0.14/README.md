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
[Приёмка PoB II со стабильным движком](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45297407858).

## PoB II v1.0.0

Выпуск собран из коммита `0ea165dc4e7b14ccbf16c5417cfac9b2ac0f120c`.
На трёх ОС проверены чистая установка, обновление старого профиля и
восстановление сборки. Затем скачанные установщики сверены с `SHA256SUMS`.
Проверка каждого скачанного установщика включает запуск, расчёт Lua,
сохранение XML и повторное открытие сборки с заметками.
[Выпуск PoB II](https://github.com/Hecatoncheir/pob2_ui/releases/tag/v1.0.0) ·
[Выпускной прогон](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45306825892).

| Платформа | SHA-256 установщика | Проверка скачанного пакета |
| --- | --- | --- |
| Windows x64 | `00c00eb28f7341da6d92f45c63b5b3d10530c8762c1c13fc8bebd2165a4c1183` | [Прогон](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45306825892/job/138177582828) |
| Linux x64 | `f206ae675f01d4cc743397bcb32b17a3a605d29cc29594803634e691ff5df7db` | [Прогон](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45315522912/job/138181323276) |
| macOS arm64 | `fc4963d180e3d5f94a1d79baf7bbb9e8d0768f1c4755b1f1cfd3ec3bdd45139d` | [Прогон](https://github.com/Hecatoncheir/pob2_ui/actions/runs/45306825892/job/138177583082) |

На чистом исполнителе Ubuntu 22.04 для первого кадра Flutter понадобились
системные библиотеки EGL и OpenGL ES. Повторная проверка использовала тот же
опубликованный установщик и установила только системные библиотеки GTK,
EGL и OpenGL ES с Xvfb.
