"""Создать или проверить эталон оригинального Lua в отдельной среде."""

import argparse
import json
import math
from pathlib import Path
import re
import subprocess
import sys
import tempfile

from reference_env import CACHE, DATA, ROOT, environment, prepare, read_json, sha256, verify


def encode(value):
    return (json.dumps(value, ensure_ascii=False, indent=2, sort_keys=True, allow_nan=False) + '\n').encode('utf-8')


def compare(expected, actual, path='$'):
    if type(expected) in (int, float) and type(actual) in (int, float):
        if not math.isclose(expected, actual, rel_tol=1e-10, abs_tol=1e-9):
            raise ValueError(f'Расхождение {path}: {expected} != {actual}')
    elif isinstance(expected, dict) and isinstance(actual, dict):
        if expected.keys() != actual.keys():
            raise ValueError('Изменён набор полей: ' + path)
        for key in expected:
            compare(expected[key], actual[key], path + '.' + key)
    elif isinstance(expected, list) and isinstance(actual, list):
        if len(expected) != len(actual):
            raise ValueError('Изменена длина списка: ' + path)
        for index, pair in enumerate(zip(expected, actual)):
            compare(*pair, path + '[' + str(index) + ']')
    elif type(expected) is not type(actual) or expected != actual:
        raise ValueError('Расхождение значения: ' + path)


def clean_report(report):
    # Имя процесса MCP не сохраняется в XML; оно не является расчётным входом.
    report['snapshot']['buildInfo'].pop('name', None)
    return report


def worker(record):
    sys.path.insert(0, str(CACHE / 'mcp'))
    from pob_poe2_mcp.api import BuildAPI
    recipe = read_json(DATA / 'fireball.recipe.json')
    with BuildAPI(pob_dir=CACHE / 'pob', mode='engine', timeout=60) as api:
        capabilities = api.capabilities
        if record:
            classes = api.call('getCatalog', {'kind': 'classes'})['entries']
            class_id = next(entry['id'] for entry in classes if entry['name'] == recipe['class'])
            api.call('createBuild', {'name': 'reference-fireball', 'classId': class_id, 'level': recipe['level']})
            api.call('addSkillGroup', {'gems': [{'id': recipe['gem'], 'level': recipe['gem_level'],
                                               'quality': recipe['gem_quality']}]})
            defaults = api.call('getBuildReport')['config']
            config = {key: False for key, value in defaults.items() if key.startswith('quest') and type(value) is bool}
            api.call('setConfig', {'values': {**config, **recipe['config']}})
            xml = api.call('exportBuild')['xml']
            xml = '\n'.join(line.rstrip() for line in xml.splitlines()) + '\n'
        else:
            xml = (DATA / 'fireball.xml').read_text('utf-8')
        api.call('loadBuild', {'xml': xml})
        report = clean_report(api.call('getBuildReport'))
    # Новый процесс не должен зависеть от состояния, созданного до экспорта.
    with BuildAPI(pob_dir=CACHE / 'pob', mode='engine', timeout=60) as api:
        api.call('loadBuild', {'xml': xml})
        compare(report, clean_report(api.call('getBuildReport')))
    print(encode({'xml': xml, 'report': report, 'capabilities': capabilities}).decode('utf-8'))


def diagnostics(stderr):
    lines = stderr.replace('\r\n', '\n').splitlines()
    # Время запуска не является диагностикой расчёта и зависит от платформы.
    return [line for line in lines if line and not re.fullmatch(r'Startup time: .*', line)]


def execute(record):
    before = verify()
    with tempfile.TemporaryDirectory(prefix='lua-', dir=CACHE) as temporary:
        command = [sys.executable, '-X', 'utf8', str(Path(__file__).resolve()), '--worker']
        if record:
            command.append('--record')
        result = subprocess.run(command, cwd=ROOT, env=environment(Path(temporary)),
                                capture_output=True, text=True, encoding='utf-8', timeout=180)
    verify()
    if result.returncode:
        raise ValueError('Не удалось выполнить Lua:\n' + result.stderr)
    answer = json.loads(result.stdout)
    xml = answer['xml'].replace('\r\n', '\n').encode('utf-8')
    report = answer['report']
    golden = {'schema': 1, 'input_sha256': sha256(xml), 'recipe_sha256': sha256((DATA / 'fireball.recipe.json').read_bytes()),
              'pob_commit': before['pob_commit'], 'mcp_commit': before['mcp_commit'],
              'tree_version': before['tree_version'], 'report': report,
              'startup_diagnostics': diagnostics(result.stderr)}
    output = CACHE / 'last-run'
    output.mkdir(exist_ok=True)
    (output / 'environment.json').write_bytes(encode({**before, 'capabilities': answer['capabilities']}))
    (output / 'fireball.lua.json').write_bytes(encode(golden))
    (output / 'stderr.txt').write_text(result.stderr, encoding='utf-8')
    if record:
        (DATA / 'fireball.xml').write_bytes(xml)
        (DATA / 'fireball.lua.json').write_bytes(encode(golden))
        (DATA / 'recorded-environment.json').write_bytes(encode(before))
    else:
        compare(read_json(DATA / 'fireball.lua.json'), golden)
    print('Lua-эталон совпал; TotalDPS =', report['stats']['TotalDPS'],
          '; предупреждений:', len(report['warnings']))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--prepare', action='store_true', help='подготовить закреплённые исходники и runtime')
    parser.add_argument('--record', action='store_true', help='явно обновить XML и эталон после проверки изменений')
    parser.add_argument('--worker', action='store_true', help=argparse.SUPPRESS)
    args = parser.parse_args()
    if args.worker:
        worker(args.record)
        return
    if args.prepare:
        prepare()
    execute(args.record)


if __name__ == '__main__':
    main()
