# Многогрупповые сборки в `0.0.5`

Примеры используют установленный оригинальный PoB под LuaJIT и одну сессию
сервера. В [эталонах](../testdata/reference/0.0.5/README.md) закреплены три
сборки, их варианты и показатели. Порядок установки описан в
[руководстве выпуска](RELEASE.md).

## Импорт, расчёт и смена группы

Запустите `pob2-engine serve --profile example` и выполните в другом терминале:

```shell
pob2-engine session create --profile example --session caster --name Sorceress --request-id create-caster
pob2-engine build import --profile example --session caster --revision 1 --request-id import --file testdata/reference/0.0.5/caster.xml
pob2-engine calculate --profile example --session caster --revision 2 --request-id initial-dps
pob2-engine preview-support --profile example --session caster --revision 3 --support "Concentrated Area" --level 1 --quality 20
```

`build.calculation.stats.TotalDPS` для Ice Nova составляет `7.84`,
`FullDPS` для действующих групп — примерно `16.4043`. Предпросмотр учитывает
поддержку только в выбранной группе и возвращает исходные и новые числа.
Сохранённая ревизия после него — `3`.

Сохраните выбор Spark в `select-group.json` и примените его к ревизии `3`:

```json
{"selection":{"main_group":2}}
```

```shell
pob2-engine build edit --profile example --session caster --revision 3 --request-id select-spark --file select-group.json
pob2-engine calculate --profile example --session caster --revision 4 --request-id spark-dps
```

`TotalDPS` теперь относится к Spark, а `FullDPS` включает те же действующие
группы. На ревизии `5` можно выбрать первый навык второго набора.

## Та же правка через MCP и gRPC

Подключите `pob2-engine mcp --attach caster --profile example` и вызовите
`edit_build` с аргументами:

```json
{
  "expected_revision": 5,
  "request_id": "select-alternate",
  "edit": {"selection": {"sets": {"skills": 2}, "main_group": 1}}
}
```

В gRPC передайте те же значения в `SessionService.ApplyBuild`: `session_id`
равен `caster`, а `edit.selection.sets.skills` — `2`. `CalculateBuild`
на ревизии `6` рассчитает Frostbolt. За сохранёнными правками можно следить
через `WatchSession`. [Контракт API и события](SESSION-API.md)

Для сравнения оружия в сборке Warrior вызовите `build compare --edit` с
правкой `{"selection":{"weapon_set":2}}`. В MCP ей соответствует
`compare_build`, в gRPC — `CompareBuild`. Исходный и предлагаемый расчёты
возвращаются вместе с числовыми разницами. При статусе `partial` прочитайте
`issues`, прежде чем использовать показатели.
