# Полный каталог и правки через оригинальный PoB

Примеры рассчитаны на 0.0.4 и PoB v0.23.1. Запустите `serve --profile main`
в отдельном терминале. При первом запуске нужен доступ к GitHub для установки
PoB; установленная версия работает офлайн. Ожидаемые числовые результаты
относятся только к приведённой минимальной сборке.
XML доступен в [исходниках примера](../testdata/reference/0.0.4/ice-nova.xml).

```shell
pob2-engine catalog --profile main --kind gems --query "Ice Nova"
pob2-engine catalog --profile main --kind gems --query "Concentrated Area"
pob2-engine catalog --profile main --kind items --query "Rusted Cuirass"
pob2-engine catalog --profile main --kind passives --query 41965
pob2-engine session create --profile main --session nova --name Nova --request-id create
pob2-engine build import --profile main --session nova --revision 1 --request-id import --file testdata/reference/0.0.4/ice-nova.xml
pob2-engine calculate --profile main --session nova --revision 2 --request-id calculate
pob2-engine preview-support --profile main --session nova --revision 3 --support "Concentrated Area" --level 1 --quality 20
```

Ice Nova возвращает `TotalDPS = 7.84`. Вариант с Concentrated Area возвращает
`TotalDPS = 10.08`, исходная сессия остаётся на ревизии 3. Для MCP подключитесь
через `mcp --profile main --attach nova`. Чтобы получить ту же способность
правкой уже импортированного Fireball, передайте `edit_build`:

```json
{
  "expected_revision": 2,
  "request_id": "ice-nova",
  "edit": {
    "gems": {
      "group_id": 1,
      "enabled": true,
      "gems": [{"id": "Ice Nova", "level": 1, "quality": 0, "enabled": true}]
    }
  }
}
```

Для `preview_support` передайте `support_id: "Concentrated Area"`,
`level: 1`, `quality: 20` и текущую `expected_revision`. gRPC использует
те же поля в `PreviewSupportRequest`. `Rapid Attacks I` с Ice Nova возвращает
`compatible: false`, `partial` и причину `pob_support_incompatible`.

Правка экипировки принимает, например,
`{"equipment":{"slot":"Body Armour","raw":"Rarity: NORMAL\nRusted Cuirass"}}`.
Правка пассивов принимает ID из каталога активного дерева. Полный пример
ветки от старта Sorceress: `4739,18845,55807,6686,51184,41965`; перед её
выделением задайте достаточный уровень персонажа. Игровые ограничения и
совместимость предметов не определяются одним наличием записи в каталоге.

Каталог и предупреждения описаны в [контракте](CATALOG.md).
Сессии и исходные ответы `request_id` сохраняются после перезапуска.
Повтор старой команды возвращает первоначальный снимок и не откатывает
текущую сборку. Сам каталог в журнал сессии не записывается; восстановление
ранее принятой правки не требует повторного поиска по новому каталогу.
