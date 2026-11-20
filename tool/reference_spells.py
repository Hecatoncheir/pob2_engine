"""Отдельная матрица Lua для заклинаний холода и молнии выпуска 0.0.2."""

import argparse
import json
from pathlib import Path
import subprocess
import sys
import tempfile

from reference_cases import STATS, make_build, select_report
from reference_env import CACHE, DATA, ROOT, environment, read_json, sha256, verify
from reference_lua import compare, encode

DIRECTORY = DATA / '0.0.2'


def stats_for(element):
    return [name.replace('Fire', element, 1) if name.startswith('Fire') and name not in {'FireResist'} else name
            for name in STATS]


def checked_report(report, case):
    selected = select_report(report, stats_for(case['element']))
    unparsed = [entry['line'] for block in selected['custom_mods'] for entry in block['unparsed']]
    if unparsed != case.get('expected_unparsed', []):
        raise ValueError('Неожиданные неразобранные свойства: ' + case['id'])
    errors = [gem['error'] for group in selected['groups'] for gem in group['gems'] if gem.get('error')]
    if errors:
        raise ValueError('Ошибки камней: ' + case['id'] + ': ' + str(errors))
    return selected


def worker(record):
    sys.path.insert(0, str(CACHE / 'mcp'))
    from pob_poe2_mcp.api import BuildAPI
    outputs = []
    with BuildAPI(pob_dir=CACHE / 'pob', mode='engine') as api:
        gems = {}
        for name in ['Frostbolt', 'Spark', 'Rapid Casting I', 'Controlled Destruction']:
            entries = api.call('getCatalog', {'query': name})['entries']
            gems[name] = next(gem['id'] for gem in entries if gem['name'] == name)
        for case in read_json(DIRECTORY / 'spells.recipe.json')['cases']:
            xml = make_build(api, case, gems) if record else (DIRECTORY / 'cases' / (case['id'] + '.xml')).read_text('utf-8')
            api.call('loadBuild', {'xml': xml})
            outputs.append({'id': case['id'], 'skill': case['skill'], 'element': case['element'], 'xml': xml,
                            'report': checked_report(api.call('getBuildReport'), case)})
    print(encode(outputs).decode('utf-8'))


def execute(record):
    state = verify()
    with tempfile.TemporaryDirectory(prefix='spells-', dir=CACHE) as home:
        command = [sys.executable, '-X', 'utf8', str(Path(__file__).resolve()), '--worker']
        if record:
            command.append('--record')
        process = subprocess.run(command, cwd=ROOT, env=environment(Path(home)), capture_output=True,
                                 encoding='utf-8', text=True, timeout=240)
    verify()
    if process.returncode:
        raise ValueError('Ошибка матрицы заклинаний Lua:\n' + process.stderr)
    output = []
    (DIRECTORY / 'cases').mkdir(exist_ok=True)
    for case in json.loads(process.stdout):
        xml = case.pop('xml').encode('utf-8')
        if record:
            (DIRECTORY / 'cases' / (case['id'] + '.xml')).write_bytes(xml)
        output.append({**case, 'input_sha256': sha256(xml)})
    golden = {'schema': 1, 'pob_commit': state['pob_commit'], 'mcp_commit': state['mcp_commit'],
              'recipe_sha256': sha256((DIRECTORY / 'spells.recipe.json').read_bytes()), 'cases': output}
    log = CACHE / 'last-run'
    log.mkdir(exist_ok=True)
    (log / 'spells.lua.json').write_bytes(encode(golden))
    (log / 'spells-environment.json').write_bytes(encode(state))
    (log / 'spells-stderr.txt').write_text(process.stderr, encoding='utf-8')
    if record:
        (DIRECTORY / 'spells.lua.json').write_bytes(encode(golden))
    else:
        compare(read_json(DIRECTORY / 'spells.lua.json'), golden)
    print('Матрица заклинаний Lua совпала:', len(output), 'вариантов')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--record', action='store_true', help='явно обновить новые XML и эталоны')
    parser.add_argument('--worker', action='store_true', help=argparse.SUPPRESS)
    args = parser.parse_args()
    if args.worker:
        worker(args.record)
    else:
        execute(args.record)


if __name__ == '__main__':
    main()
