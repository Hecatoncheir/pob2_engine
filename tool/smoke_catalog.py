"""Проверить полный каталог оригинального PoB и правки через нативный бинарник."""

import math


def require(condition, message):
    if not condition:
        raise ValueError(message)


def catalog_page(scenario, lock, kind, query='', limit=100):
    result = scenario.run('catalog', *scenario.connection, '--kind', kind,
                          '--query', query, '--limit', limit)
    from_mcp, = scenario.mcp_tools('catalog', [('get_catalog', dict(kind=kind, query=query, limit=limit))])
    require(result == from_mcp, 'Разные страницы каталога CLI/gRPC и MCP')
    require(result['data_id'] == 'pob-lua-' + lock['tag'] and result['pob_commit'] == lock['commit'],
            'Каталог не соответствует оригинальному расчётчику')
    require(result['total'] > 0 and len(result['data_sha256']) == 64, 'Пустой каталог или нет происхождения')
    return result


def apply_edit(scenario, snapshot, request, edit):
    result, = scenario.mcp_tools('catalog', [('edit_build', dict(expected_revision=snapshot['revision'], request_id=request, edit=edit))])
    require(result['revision'] == snapshot['revision'] + 1, 'Правка не сохранила новую ревизию')
    return result


def preview_support(scenario, snapshot, support, quality, code=0):
    arguments = dict(expected_revision=snapshot['revision'], support_id=support, level=1, quality=quality)
    cli = scenario.run('preview-support', *scenario.connection, '--session', 'catalog',
                       '--revision', snapshot['revision'], '--support', support, '--level', '1', '--quality', quality, code=code)
    mcp, = scenario.mcp_tools('catalog', [('preview_support', arguments)])
    require(cli == mcp, 'Разный предпросмотр CLI/gRPC и MCP')
    require(scenario.snapshot('catalog') == snapshot, 'Предпросмотр изменил сессию')
    return cli


def check_new_skill(scenario, lock, verify):
    previous = scenario.snapshot('catalog')
    edit = dict(gems=dict(group_id=1, enabled=True,
                         gems=[dict(id='Ice Nova', level=1, quality=0, enabled=True)]))
    edited = apply_edit(scenario, previous, 'ice-nova', edit)
    calculated = scenario.calculate('catalog', edited['revision'], 'ice-nova-calculate')
    verify(calculated['build']['calculation'], 'Ice Nova', 7.84, lock)
    preview = preview_support(scenario, calculated, 'Concentrated Area', 20)
    require(preview['compatible'] and preview['status'] == 'complete', 'Совместимая поддержка отклонена')
    verify(preview['comparison']['after'], 'Ice Nova', 10.08, lock)
    incompatible = preview_support(scenario, calculated, 'Rapid Attacks I', 0, code=4)
    require(not incompatible['compatible'] and incompatible['status'] == 'partial', 'Несовместимая поддержка не отмечена')
    return edited, calculated, preview, edit


def check_equipment_and_passives(scenario, snapshot, lock):
    raw = 'Rarity: NORMAL\nRusted Cuirass'
    equipped = apply_edit(scenario, snapshot, 'body-armour', dict(equipment=dict(slot='Body Armour', raw=raw)))
    levelled = apply_edit(scenario, equipped, 'level-ten', dict(level=dict(level=10)))
    nodes = [4739, 18845, 55807, 6686, 51184, 41965]
    allocated = apply_edit(scenario, levelled, 'new-passive', dict(passives=dict(nodes=nodes)))
    require(41965 in allocated['build']['model']['nodes'], 'Новый пассив потерян')
    result = scenario.calculate('catalog', allocated['revision'], 'equipment-passives-calculate')
    calculation = result['build']['calculation']
    require(calculation['status'] == 'complete' and calculation['data_id'] == 'pob-lua-' + lock['tag'], 'Неполный расчёт брони и пассивов')
    require(calculation['stats']['Armour'] > 0 and math.isfinite(calculation['stats']['TotalDPS']), 'Броня или урон не рассчитаны')
    require(any(item['raw'] == raw for item in result['build']['model']['items']), 'Изменён исходный текст предмета')
    return result


def apply_previewed_support(scenario, snapshot, preview, lock, verify):
    edit = dict(add_support=dict(support_id=preview['support']['id'], level=1, quality=20))
    applied = apply_edit(scenario, snapshot, 'append-support', edit)
    expected = preview['comparison']['after']
    require(applied['build']['input_sha256'] == expected['input_sha256'], 'Правка отличается от предпросмотра')
    require(not applied['build'].get('calculation'), 'После правки сохранён прежний расчёт')
    saved = scenario.calculate('catalog', applied['revision'], 'appended-support-calculate')
    verify(saved['build']['calculation'], 'Ice Nova', 10.08, lock)
    return applied, saved, edit


def check_full_catalog(scenario, lock, verify):
    counts = {}
    for kind, query in [('gems', 'Ice Nova'), ('classes', 'Sorceress'), ('items', 'Rusted Cuirass'), ('passives', '41965')]:
        counts[kind] = catalog_page(scenario, lock, kind, limit=1)['total']
        catalog_page(scenario, lock, kind, query)
    require(counts['gems'] > 100 and counts['items'] > 100 and counts['passives'] > 1000, 'Каталог остался ограниченным набором Go')
    edited, calculated, preview, edit = check_new_skill(scenario, lock, verify)
    applied, saved, support_edit = apply_previewed_support(scenario, calculated, preview, lock, verify)
    final = check_equipment_and_passives(scenario, saved, lock)
    scenario.stop()
    scenario.start('lua')
    replay, = scenario.mcp_tools('catalog', [('edit_build', dict(expected_revision=2, request_id='ice-nova', edit=edit))])
    support_replay, = scenario.mcp_tools('catalog', [('edit_build', dict(
        expected_revision=calculated['revision'], request_id='append-support', edit=support_edit))])
    require(replay == edited and support_replay == applied and scenario.snapshot('catalog') == final,
            'После перезапуска потерян ответ или изменена сессия')
    return dict(counts=counts, new_skill_dps=7.84, supported_dps=10.08, support_quality=20,
                incompatible_support_reported=True, applied_preview_matches=True,
                equipment_and_passives=True, restart_exact_retry=True)
