"""Воспроизводимый экспорт выбранных данных; Go использует встроенный набор."""

import argparse
from pathlib import Path
import subprocess
import sys
import tempfile

from reference_env import CACHE, ROOT, environment, read_json, sha256, verify
from reference_lua import encode

OUTPUT = ROOT / 'internal/gamedata/files'
INPUTS = ['src/Data/Gems.lua', 'src/Data/Skills/act_int.lua', 'src/Data/Skills/sup_int.lua',
          'src/Data/Bases/ring.lua', 'src/Data/Bases/amulet.lua', 'src/Data/Misc.lua',
          'src/TreeData/0_5/tree.json', 'src/Modules/CalcTools.lua', 'src/Modules/Data.lua']


def export():
    state = verify()
    with tempfile.TemporaryDirectory(prefix='data-', dir=CACHE) as home:
        process = subprocess.run([sys.executable, '-X', 'utf8', str(CACHE / 'mcp/pob_poe2_mcp/runner.py'),
                                  '--pob-dir', str(CACHE / 'pob'), '--script', str(ROOT / 'tool/export_data.lua')],
                                 cwd=ROOT, env=environment(Path(home)), timeout=120,
                                 capture_output=True, text=True, encoding='utf-8')
    verify()
    if process.returncode:
        raise ValueError('Ошибка извлечения данных:\n' + process.stderr)
    import json
    data = json.loads(process.stdout)
    for key, identifier in [('classes', 'id'), ('gems', 'id'), ('items', 'name'), ('nodes', 'id')]:
        data[key].sort(key=lambda entry: entry[identifier])
    for gem in data['gems']:
        gem['constant_stats'] = [{'name': name, 'value': value} for name, value in gem['constant_stats']]
    # Связи дерева не направлены; исходный JSON записывает каждую дугу один раз.
    nodes = {entry['id']: entry for entry in data['nodes']}
    for node in data['nodes']:
        for neighbor in list(node['neighbors']):
            nodes[neighbor]['neighbors'].append(node['id'])
    for node in data['nodes']:
        node['neighbors'] = sorted(set(node['neighbors']))
    catalog = encode(data)
    sources = read_json(ROOT / 'testdata/reference/sources.lock.json')
    manifest = {'schema': 1, 'id': 'pob2-0.0.1-fireball-v1', 'game_version': None,
                'tree_version': state['tree_version'], 'pob_commit': state['pob_commit'],
                'files': {'catalog.json': sha256(catalog)},
                'exporters': {name: sha256((ROOT / name).read_bytes()) for name in ['tool/export_data.lua', 'tool/export_data.py']},
                'origin': {'repository': sources['pob']['repository'], 'commit': state['pob_commit'],
                           'files': {name: sha256((CACHE / 'pob' / name).read_bytes()) for name in INPUTS}},
                'rights': {'code_license': 'MIT', 'upstream_notice': '.licenses/reference/pob-LICENSE.md',
                           'game_data_copyright': 'Skill, gem and item data (c) Grinding Gear Games',
                           'changes': 'Selected fields; JSON schema, sorting and undirected node links; no images or descriptions'}}
    return {'catalog.json': catalog, 'manifest.json': encode(manifest)}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--write', action='store_true', help='обновить встроенный набор после проверки diff')
    args = parser.parse_args()
    files = export()
    for name, content in files.items():
        path = OUTPUT / name
        if args.write:
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(content)
        elif not path.is_file() or path.read_bytes() != content:
            raise ValueError('Устаревший набор данных: ' + name)
    print('Набор данных воспроизведён:', len(files), 'файла')


if __name__ == '__main__':
    main()
