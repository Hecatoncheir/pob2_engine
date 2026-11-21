"""Матрицы Lua для заклинаний, базовых классов и выбора групп выпуска 0.0.2."""

import argparse
import copy
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import xml.etree.ElementTree as ET

from reference_cases import STATS, make_build, select_report
from reference_env import CACHE, DATA, ROOT, environment, read_json, sha256, verify
from reference_lua import compare, encode

DIRECTORY = DATA / '0.0.2'


def stats_for(element):
    stats = list(STATS)
    for index, name in enumerate(stats):
        # Сопротивление персонажа общее; меняется только стихия показателей урона.
        if name.startswith('Fire') and name != 'FireResist':
            stats[index] = element + name.removeprefix('Fire')
    return stats


def checked_report(report, case):
    selected = select_report(report, stats_for(case['element']))
    unparsed = [entry['line'] for block in selected['custom_mods'] for entry in block['unparsed']]
    if unparsed != case.get('expected_unparsed', []):
        raise ValueError('Неожиданные неразобранные свойства: ' + case['id'])
    errors = [gem['error'] for group in selected['groups'] for gem in group['gems'] if gem.get('error')]
    if errors:
        raise ValueError('Ошибки камней: ' + case['id'] + ': ' + str(errors))
    return selected


def matrix_paths(matrix):
    return DIRECTORY if matrix == 'spells' else DIRECTORY / matrix


def make_group_build(api, case, gems):
    make_build(api, {**case, 'skill': 'Fireball'}, gems)
    for name in ['Frostbolt', 'Spark']:
        api.call('addSkillGroup', {'gems': [{'id': gems[name], 'level': 1, 'quality': 0}]})
    root = ET.fromstring(api.call('exportBuild')['xml'])
    root.find('Build').set('mainSocketGroup', str(case['main_group']))
    skills = root.find('Skills')
    primary = skills.find('SkillSet')
    groups = primary.findall('Skill')
    alternate = ET.SubElement(skills, 'SkillSet', id='2', title='alternate')
    alternate.append(copy.deepcopy(groups[2]))
    for index, group in enumerate(groups, 1):
        group.set('enabled', str(index == case['main_group']).lower())
    skills.set('activeSkillSet', str(case['skill_set']))
    config = root.find('Config')
    other_config = copy.deepcopy(config.find('ConfigSet'))
    other_config.set('id', '2')
    custom = other_config.find("Input[@name='customMods']")
    if custom is None:
        custom = ET.SubElement(other_config, 'Input', name='customMods')
    custom.set('string', '100% increased Spell Damage')
    config.append(other_config)
    config.set('activeConfigSet', str(case['config_set']))
    api.call('loadBuild', {'xml': ET.tostring(root, encoding='unicode')})
    return '\n'.join(line.rstrip() for line in api.call('exportBuild')['xml'].splitlines()) + '\n'


def catalog_gems(api):
    gems = {}
    for name in ['Fireball', 'Frostbolt', 'Spark', 'Rapid Casting I', 'Controlled Destruction']:
        entries = api.call('getCatalog', {'query': name})['entries']
        gems[name] = next(gem['id'] for gem in entries if gem['name'] == name)
    return gems


def class_metadata(case, report):
    if 'class_id' not in case:
        return {}
    if report['allocated_nodes'] != [case['start_node']]:
        raise ValueError('Не совпал старт класса: ' + case['id'])
    return {key: case[key] for key in ['class_id', 'class_name', 'start_node']}


def collect_case(api, case, xml):
    api.call('loadBuild', {'xml': xml})
    report = checked_report(api.call('getBuildReport'), case)
    return {'id': case['id'], 'skill': case['skill'], 'element': case['element'], 'xml': xml,
            'report': report, **class_metadata(case, report)}


def worker(record, matrix):
    sys.path.insert(0, str(CACHE / 'mcp'))
    from pob_poe2_mcp.api import BuildAPI
    directory = matrix_paths(matrix)
    builder = make_group_build if matrix == 'groups' else make_build
    outputs = []
    with BuildAPI(pob_dir=CACHE / 'pob', mode='engine') as api:
        gems = catalog_gems(api)
        recipe = read_json(directory / (matrix + '.recipe.json'))
        for case in recipe['cases']:
            if record:
                xml = builder(api, case, gems)
            else:
                xml = (directory / 'cases' / (case['id'] + '.xml')).read_text('utf-8')
            outputs.append(collect_case(api, case, xml))
    print(encode(outputs).decode('utf-8'))


def execute(record, matrix):
    state = verify()
    directory = matrix_paths(matrix)
    with tempfile.TemporaryDirectory(prefix='spells-', dir=CACHE) as home:
        command = [sys.executable, '-X', 'utf8', str(Path(__file__).resolve()), '--worker', '--matrix', matrix]
        if record:
            command.append('--record')
        process = subprocess.run(command, cwd=ROOT, env=environment(Path(home)), capture_output=True,
                                 encoding='utf-8', text=True, timeout=240)
    verify()
    if process.returncode:
        raise ValueError('Ошибка матрицы Lua:\n' + process.stderr)
    output = []
    (directory / 'cases').mkdir(exist_ok=True)
    for case in json.loads(process.stdout):
        xml = case.pop('xml').encode('utf-8')
        if record:
            (directory / 'cases' / (case['id'] + '.xml')).write_bytes(xml)
        output.append({**case, 'input_sha256': sha256(xml)})
    golden = {'schema': 1, 'pob_commit': state['pob_commit'], 'mcp_commit': state['mcp_commit'],
              'recipe_sha256': sha256((directory / (matrix + '.recipe.json')).read_bytes()), 'cases': output}
    log = CACHE / 'last-run'
    log.mkdir(exist_ok=True)
    (log / (matrix + '.lua.json')).write_bytes(encode(golden))
    (log / (matrix + '-environment.json')).write_bytes(encode(state))
    (log / (matrix + '-stderr.txt')).write_text(process.stderr, encoding='utf-8')
    if record:
        (directory / (matrix + '.lua.json')).write_bytes(encode(golden))
    else:
        compare(read_json(directory / (matrix + '.lua.json')), golden)
    print('Матрица Lua совпала:', matrix, len(output), 'вариантов')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--record', action='store_true', help='явно обновить новые XML и эталоны')
    parser.add_argument('--worker', action='store_true', help=argparse.SUPPRESS)
    parser.add_argument('--matrix', choices=['spells', 'classes', 'groups'], default='spells', help='набор эталонных случаев')
    args = parser.parse_args()
    if args.worker:
        worker(args.record, args.matrix)
    else:
        execute(args.record, args.matrix)


if __name__ == '__main__':
    main()
