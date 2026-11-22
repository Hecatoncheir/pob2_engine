"""Записать или воспроизвести эталоны многогрупповых сборок оригинального PoB."""

import json
from pathlib import Path
import subprocess
import sys
import tempfile

from reference_env import CACHE, DATA, ROOT, environment, read_json, sha256, verify
from reference_lua import compare, encode

DIRECTORY = DATA / '0.0.5'
STATS = ('TotalDPS', 'FullDPS', 'Life', 'EnergyShield', 'Armour', 'FireResist',
         'ColdResist', 'LightningResist', 'ChaosResist', 'IgniteDPS', 'ImpaleDPS', 'Spirit')


def reference_case(api, path):
    xml = path.read_bytes()
    api.call('loadBuild', {'xml': xml.decode('utf-8')})
    report = api.call('getBuildReport')
    info = report['snapshot']['buildInfo']
    selected = info['skills']
    skill = selected['entries'][selected['selected'] - 1]['label']
    unsupported = [line['line'] for block in report['customMods'] if block['enabled'] for line in block['unparsed']]
    return dict(input_sha256=sha256(xml), skill=skill, main_group=info['groups']['selected'],
                active_sets=report['activeSets'], weapon_set=report['weaponSet'],
                enabled_groups=[group['enabled'] for group in report['groups']],
                gem_errors=[gem['error'] for group in report['groups'] if group['enabled']
                            for gem in group['gems'] if gem.get('error') and gem['enabled']],
                stats={name: report['stats'].get(name) for name in STATS},
                warnings=report['warnings'], unsupported=unsupported)


def worker():
    sys.path.insert(0, str(CACHE / 'mcp'))
    from pob_poe2_mcp.api import BuildAPI

    with BuildAPI(pob_dir=CACHE / 'pob', mode='engine') as api:
        cases = {path.stem: reference_case(api, path) for path in sorted(DIRECTORY.glob('*.xml'))}
    print(encode(cases).decode('utf-8'), end='')


def collect_cases():
    with tempfile.TemporaryDirectory(prefix='multigroup-reference-', dir=CACHE) as home:
        process = subprocess.run([sys.executable, '-X', 'utf8', str(Path(__file__).resolve()), '--worker'],
                                 cwd=ROOT, env=environment(Path(home)), capture_output=True,
                                 text=True, encoding='utf-8', timeout=120)
    if process.returncode:
        raise ValueError('Ошибка эталона многогрупповых сборок:\n' + process.stderr)
    return json.loads(process.stdout)


def execute(record):
    state = verify()
    outcome = dict(pob_commit=state['pob_commit'], mcp_commit=state['mcp_commit'],
                   cases=collect_cases(), schema=1)
    verify()
    golden = DIRECTORY / 'golden.json'
    if record:
        golden.write_bytes(encode(outcome))
    else:
        compare(read_json(golden), outcome)
    output = CACHE / 'last-run'
    output.mkdir(exist_ok=True)
    (output / 'multigroup.lua.json').write_bytes(encode(outcome))
    print('Эталоны многогрупповых сборок Lua:', len(outcome['cases']))


if __name__ == '__main__':
    worker() if '--worker' in sys.argv else execute('--record' in sys.argv)
