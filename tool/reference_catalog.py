"""Воспроизвести примеры нового каталога независимым адаптером оригинального Lua."""

import json
from pathlib import Path
import subprocess
import sys
import tempfile

from reference_env import CACHE, DATA, ROOT, environment, read_json, sha256, verify
from reference_lua import compare, encode

DIRECTORY = DATA / '0.0.4'


def worker():
    sys.path.insert(0, str(CACHE / 'mcp'))
    from pob_poe2_mcp.api import BuildAPI
    cases = []
    with BuildAPI(pob_dir=CACHE / 'pob', mode='engine') as api:
        for case in read_json(DIRECTORY / 'golden.json')['cases']:
            xml = (DIRECTORY / (case['id'] + '.xml')).read_bytes()
            if sha256(xml) != case['input_sha256']:
                raise ValueError('Изменён эталон XML: ' + case['id'])
            api.call('loadBuild', {'xml': xml.decode('utf-8')})
            report = api.call('getBuildReport')
            cases.append(dict(id=case['id'], input_sha256=sha256(xml),
                              stats=report['stats'], warnings=report['warnings'], groups=report['groups']))
    print(encode(cases).decode('utf-8'))


def execute():
    state = verify()
    golden = read_json(DIRECTORY / 'golden.json')
    if (golden['pob_commit'], golden['mcp_commit']) != (state['pob_commit'], state['mcp_commit']):
        raise ValueError('Версии эталонов каталога не соответствуют закреплённым исходникам')
    with tempfile.TemporaryDirectory(prefix='catalog-reference-', dir=CACHE) as home:
        process = subprocess.run([sys.executable, '-X', 'utf8', str(Path(__file__).resolve()), '--worker'],
                                 cwd=ROOT, env=environment(Path(home)), capture_output=True,
                                 text=True, encoding='utf-8', timeout=120)
    verify()
    if process.returncode:
        raise ValueError('Ошибка эталона каталога:\n' + process.stderr)
    actual = json.loads(process.stdout)
    compare(golden['cases'], actual)
    output = CACHE / 'last-run'
    output.mkdir(exist_ok=True)
    (output / 'catalog.lua.json').write_bytes(encode(actual))
    (output / 'catalog-environment.json').write_bytes(encode(state))
    print('Эталоны нового каталога Lua совпали:', len(actual))


if __name__ == '__main__':
    worker() if '--worker' in sys.argv else execute()
