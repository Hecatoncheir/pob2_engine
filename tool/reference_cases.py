"""Матрица расчётов Lua для проверки выбранного покрытия Go."""

import argparse
import json
from pathlib import Path
import subprocess
import sys
import tempfile

from reference_env import CACHE, DATA, ROOT, environment, read_json, sha256, verify
from reference_lua import compare, encode

STATS = ['Str', 'Dex', 'Int', 'Life', 'Mana', 'ManaRegen', 'ManaCost', 'ManaPerSecondCost',
         'CastRate', 'Speed', 'CritChance', 'CritMultiplier', 'CritEffect', 'AverageHit', 'TotalDPS',
         'FireResist', 'ColdResist', 'LightningResist', 'ChaosResist',
         'FireMinBase', 'FireMaxBase', 'FireHitAverage', 'FireCritAverage',
         'FireStoredHitMin', 'FireStoredHitMax', 'FireStoredHitAvg', 'FireStoredCombinedAvg']
FIXTURES = DATA / 'cases'


def make_build(api, case, gems):
    api.call('createBuild', {'classId': 7, 'level': case['level'], 'name': case['id']})
    group = [{'id': gems['Fireball'], 'level': case['gem_level'], 'quality': 0}]
    group += [{'id': gems[name], 'level': 1} for name in case.get('supports', [])]
    api.call('addSkillGroup', {'gems': group})
    defaults = api.call('getBuildReport')['config']
    config = {key: False for key, value in defaults.items() if key.startswith('quest') and type(value) is bool}
    api.call('setConfig', {'values': {**config, 'enemyIsBoss': 'None', 'resistancePenalty': 0, **case.get('config', {})}})
    api.call('setCustomMods', {'text': case.get('custom_mods', '')})
    for node in case.get('nodes', []):
        api.call('toggleNode', {'id': node})
    for item in case.get('items', []):
        inspected = api.call('inspectItem', item)
        if inspected['item']['unsupported'] or not inspected['compatible']:
            raise ValueError('Неподдержанный Lua предмет в примере ' + case['id'])
        api.call('addItem', item)
    xml = api.call('exportBuild')['xml']
    return '\n'.join(line.rstrip() for line in xml.splitlines()) + '\n'


def select_report(report):
    return {'stats': {name: report['stats'][name] for name in STATS},
            'warnings': report['warnings'], 'groups': report['groups'], 'items': report['items'],
            'custom_mods': report['customMods'],
            'allocated_nodes': report['allocatedNodes'], 'config': report['config'],
            'slots': [slot for slot in report['slots'] if slot['itemId'] != 0]}


def worker(record):
    sys.path.insert(0, str(CACHE / 'mcp'))
    from pob_poe2_mcp.api import BuildAPI
    recipe = read_json(DATA / 'cases.recipe.json')
    outputs = []
    with BuildAPI(pob_dir=CACHE / 'pob', mode='engine') as api:
        gems = {}
        for name in ['Fireball', 'Rapid Casting I', 'Controlled Destruction']:
            entries = api.call('getCatalog', {'query': name})['entries']
            gems[name] = next(gem['id'] for gem in entries if gem['name'] == name)
        for case in recipe['cases']:
            xml = make_build(api, case, gems) if record else (FIXTURES / (case['id'] + '.xml')).read_text('utf-8')
            api.call('loadBuild', {'xml': xml})
            report = select_report(api.call('getBuildReport'))
            unparsed = [entry['line'] for block in report['custom_mods'] for entry in block['unparsed']]
            if sorted(unparsed) != sorted(case.get('expected_unparsed', [])):
                raise ValueError('Неожиданные неразобранные свойства Lua: ' + case['id'] + ': ' + str(unparsed))
            errors = [gem['error'] for group in report['groups'] for gem in group['gems'] if gem.get('error')]
            if errors:
                raise ValueError('Ошибки камней Lua: ' + case['id'] + ': ' + str(errors))
            outputs.append({'id': case['id'], 'xml': xml, 'report': report})
    print(encode(outputs).decode('utf-8'))


def execute(record):
    state = verify()
    with tempfile.TemporaryDirectory(prefix='cases-', dir=CACHE) as home:
        command = [sys.executable, '-X', 'utf8', str(Path(__file__).resolve()), '--worker']
        if record:
            command.append('--record')
        process = subprocess.run(command, cwd=ROOT, env=environment(Path(home)), capture_output=True,
                                 encoding='utf-8', text=True, timeout=240)
    verify()
    if process.returncode:
        raise ValueError('Ошибка матрицы Lua:\n' + process.stderr)
    output = []
    FIXTURES.mkdir(exist_ok=True)
    for case in json.loads(process.stdout):
        xml = case.pop('xml').encode('utf-8')
        if record:
            (FIXTURES / (case['id'] + '.xml')).write_bytes(xml)
        output.append({**case, 'input_sha256': sha256(xml)})
    golden = {'schema': 1, 'pob_commit': state['pob_commit'], 'mcp_commit': state['mcp_commit'],
              'recipe_sha256': sha256((DATA / 'cases.recipe.json').read_bytes()), 'cases': output}
    log = CACHE / 'last-run'
    log.mkdir(exist_ok=True)
    (log / 'cases.lua.json').write_bytes(encode(golden))
    (log / 'cases-environment.json').write_bytes(encode(state))
    (log / 'cases-stderr.txt').write_text(process.stderr, encoding='utf-8')
    if record:
        (DATA / 'cases.lua.json').write_bytes(encode(golden))
    else:
        compare(read_json(DATA / 'cases.lua.json'), golden)
    print('Матрица Lua совпала:', len(output), 'вариантов')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--record', action='store_true', help='явно обновить фикстуры и числовые эталоны')
    parser.add_argument('--worker', action='store_true', help=argparse.SUPPRESS)
    args = parser.parse_args()
    if args.worker:
        worker(args.record)
    else:
        execute(args.record)


if __name__ == '__main__':
    main()
