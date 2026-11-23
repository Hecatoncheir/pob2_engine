# Эталон примерки доспеха

Сборка — [Воин с двумя наборами экипировки](../0.0.5/attacker.xml).
Ревизии PoB `v0.23.1` и MCP закреплены в [sources.lock.json](../sources.lock.json).
В активном наборе заменяется `Body Armour`, второй набор и оружие сохраняются.

```text
Rarity: RARE
Release Ward
Iron Cuirass
--------
+30 to maximum Life
+20% to Fire Resistance
```

Для независимой проверки в закреплённом Lua загрузите сборку через
`loadBuild`, прочитайте `getBuildReport`, примените этот текст через
`equipItem` со слотом `Body Armour` и снова прочитайте `getBuildReport`.
Условия боя и активное умение остаются прежними. Допуск сравнения `1e-8`.

| Показатель PoB | До | После |
| --- | ---: | ---: |
| `Life` | 321 | 353 |
| `Armour` | 45 | 115 |
| `FireResist` | 10 | 30 |
| `TotalDPS` | 2.3287313871353 | 2.3287313871353 |

Исходный отчёт без предупреждений. После примерки PoB сообщает о нехватке
силы для `Release Ward, Iron Cuirass`. Дополнительная строка
`Unknown modifier for the equipment test` должна появиться в диагностике
движка как `unsupported_modifier` в `Items/Body Armour`.

Процессный тест [CLI, MCP и gRPC](../../../internal/integration/equipment_process_test.go)
сверяет этот эталон, состояние сессии и хеш входа. Используйте ту же подготовку
Lua и переменные окружения, что в [задании CI](../../../.github/workflows/ci.yml).
