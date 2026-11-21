# Исходные измерения 0.0.2 на Linux

[Запуск 44868204700](https://github.com/Hecatoncheir/pob2_engine/actions/runs/44868204700)
для commit `edafd1703e54a5cc3bb95a7faff8a661a5694367` полностью прошёл.
Артефакт `calculation-performance-linux` скачан после завершения CI.
Файлы ниже перенесены без изменения байтов; хеши входов проверены.
Завершающие пробелы в строке CPU исходного вывода Go сохранены для совпадения
SHA-256. Правило проверки пробелов ограничено тремя файлами вывода benchmark.

- [spells.txt](spells.txt) — три последовательных расчёта, по пять повторений.
- [support-preview.txt](support-preview.txt) — шесть предпросмотров, по пять повторений.
- [benchmark.txt](benchmark.txt) — прежний последовательный/параллельный Fireball,
  по три повторения; эти строки не входят в новую таблицу девяти сценариев.
- [go.txt](go.txt), [environment.json](environment.json) — версия и параметры Go.
- [cpu.txt](cpu.txt), [kernel.txt](kernel.txt), [os-release.txt](os-release.txt) — машина и ОС.
- [runtime.json](runtime.json) — значения GOMAXPROCS/GOGC/GOMEMLIMIT;
  `null` означает отсутствие переменной. Фактический GOMAXPROCS=4 виден в именах замеров.
- [commit.txt](commit.txt), [measured-at.txt](measured-at.txt) — исходный commit и время UTC перед серией.
- [inputs.sha256](inputs.sha256) — SHA-256 трёх XML, каталога и манифеста набора.

[summary.json](summary.json) добавлен при разборе артефакта: все 45 точек,
медианы/минимумы/максимумы времени и выделений, хеши исходных файлов.
[Методика, таблица и границы](../../../docs/PERFORMANCE-0.0.2.md) объясняют единицы:
медиана средних повторений не является p50 запросов, B/op не является памятью процесса.
Сопоставление скорости с Lua не выполнялось.
