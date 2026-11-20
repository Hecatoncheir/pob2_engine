"""Проверить распакованный релиз на настоящей целевой ОС без среды разработки."""

import argparse
from datetime import datetime, timezone
import hashlib
import json
import math
import os
from pathlib import Path
import platform
import queue
import subprocess
import tempfile
import threading
import zipfile


def require(condition, message):
    if not condition:
        raise ValueError(message)


def unpack(archive, destination, tag):
    with zipfile.ZipFile(archive) as source:
        names = source.namelist()
        require(len(names) == len(set(names)), 'Повтор имени в архиве')
        manifest = json.loads(source.read('PACKAGE.json'))
        require(manifest['schema'] == 1 and manifest['tag'] == tag, 'Неверный манифест')
        require(set(names) == set(manifest['files']) | {'PACKAGE.json'}, 'Неполный состав архива')
        for name, digest in manifest['files'].items():
            path = (destination / name).resolve()
            require(path.is_relative_to(destination), 'Путь вне архива')
            data = source.read(name)
            require(hashlib.sha256(data).hexdigest() == digest, 'SHA-256 не совпадает: ' + name)
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(data)
    executable = destination / manifest['executable']
    executable.chmod(0o755)
    return manifest, executable


def line_reader(stream, messages):
    try:
        for line in stream:
            messages.put(json.loads(line))
    except Exception as error:
        messages.put(error)
    finally:
        messages.put(EOFError('Поток завершён'))


def receive(messages):
    value = messages.get(timeout=20)
    if isinstance(value, Exception):
        raise value
    return value


class Scenario:
    def __init__(self, binary, root):
        self.binary, self.root = binary, root
        self.environment = {key: value for key, value in os.environ.items()
                            if key.upper() not in {'PATH', 'HOME', 'USERPROFILE', 'APPDATA', 'LOCALAPPDATA', 'XDG_CONFIG_HOME'}
                            and not key.upper().startswith(('POB', 'PYTHON'))}
        self.environment.update(PATH='', HOME=str(root), USERPROFILE=str(root), APPDATA=str(root),
                                LOCALAPPDATA=str(root), XDG_CONFIG_HOME=str(root), POB_DIR=str(root / 'missing-pob'))
        self.server = None
        self.connection = []
        self.steps = []

    def run(self, *args, code=0, raw=False):
        result = subprocess.run([str(self.binary), *map(str, args)], cwd=self.root,
                                env=self.environment, capture_output=True, encoding='utf-8', timeout=20)
        require(result.returncode == code, f'{args[0]}: выход {result.returncode}, stderr={result.stderr}')
        return result.stdout if raw else json.loads(result.stdout)

    def start(self, profile):
        self.server = subprocess.Popen([str(self.binary), 'serve', '--profile', profile,
                                        '--data-dir', str(self.root / 'data'), '--runtime-dir', str(self.root / 'run')],
                                       cwd=self.root, env=self.environment, stdout=subprocess.PIPE,
                                       stderr=subprocess.DEVNULL, text=True, encoding='utf-8')
        messages = queue.Queue()
        threading.Thread(target=line_reader, args=(self.server.stdout, messages), daemon=True).start()
        ready = receive(messages)
        require(ready['storage_mode'] == 'disk', 'Сервер не сохраняет сессию')
        self.connection = ['--profile', profile, '--data-dir', str(self.root / 'data'),
                           '--runtime-dir', str(self.root / 'run')]

    def stop(self):
        self.run('shutdown', *self.connection)
        require(self.server.wait(timeout=10) == 0, 'Ошибка завершения сервера')
        self.server = None

    def close(self):
        if self.server is not None:
            self.server.kill()
            self.server.wait(timeout=10)

    def mcp_edit(self):
        process = subprocess.Popen([str(self.binary), 'mcp', '--attach', 'main', *self.connection],
                                   cwd=self.root, env=self.environment, stdin=subprocess.PIPE,
                                   stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, text=True, encoding='utf-8')
        messages = queue.Queue()
        threading.Thread(target=line_reader, args=(process.stdout, messages), daemon=True).start()

        def send(message):
            process.stdin.write(json.dumps(dict(jsonrpc='2.0', **message)) + '\n')
            process.stdin.flush()

        try:
            send({'id': 1, 'method': 'initialize', 'params': {'protocolVersion': '2025-11-25',
                  'capabilities': {}, 'clientInfo': {'name': 'release-smoke', 'version': '1'}}})
            require('result' in receive(messages), 'MCP initialize не завершён')
            send({'method': 'notifications/initialized'})
            send({'id': 2, 'method': 'tools/call', 'params': {'name': 'edit_build', 'arguments': {
                'expected_revision': 2, 'request_id': 'damage', 'edit': {'config': {'name': 'customMods',
                'value': {'type': 'string', 'value': '100% increased Spell Damage'}}}}}})
            response = receive(messages)
            require(response.get('id') == 2 and 'result' in response, 'Нет ответа MCP')
            result = response['result']
            require(not result.get('isError') and result['structuredContent']['revision'] == 3, 'MCP не изменил сборку')
            process.stdin.close()
            require(process.wait(timeout=10) == 0, 'MCP не завершился по EOF')
        finally:
            if process.poll() is None:
                process.kill()
                process.wait(timeout=10)

    def execute(self, package_root, tag):
        version = self.run('--version')
        require(version['version'] == tag and version['protocol_version'] == 'pob2.v1', 'Версия бинарника не соответствует тегу')
        fixture = package_root / 'testdata/reference/fireball.xml'
        direct = self.run('calculate', '--file', fixture)
        require(direct['status'] == 'complete' and math.isclose(direct['stats']['TotalDPS'], 8.9166666666667, abs_tol=1e-9),
                'Одноразовый расчёт не совпал с Lua')
        self.steps.append('standalone_without_runtime')
        self.start('smoke')
        self.run('session', 'create', '--session', 'main', '--name', 'Fireball', '--request-id', 'create', *self.connection)
        self.run('build', 'import', '--session', 'main', '--revision', 1, '--request-id', 'import', '--file', fixture, *self.connection)
        self.mcp_edit()
        calculate = ['calculate', '--session', 'main', '--revision', 3, '--request-id', 'calculate']
        saved = self.run(*calculate, *self.connection)
        require(saved['revision'] == 4 and saved['build']['calculation']['status'] == 'complete', 'Расчёт не сохранён')
        require(math.isclose(saved['build']['calculation']['stats']['TotalDPS'], 17.8333333333333, abs_tol=1e-9), 'Правка MCP не учтена')
        self.steps.append('cli_mcp_grpc_shared_calculation')
        edit = self.root / 'edit.json'
        edit.write_text('{"config":{"name":"customMods","value":{"type":"string","value":"200% increased Spell Damage"}}}', 'utf-8')
        compared = self.run('build', 'compare', '--session', 'main', '--revision', 4, '--edit', '--file', edit, *self.connection)
        require(compared['comparison']['after']['stats']['TotalDPS'] > compared['comparison']['before']['stats']['TotalDPS'], 'Сравнение не работает')
        require(self.run('session', 'get', '--session', 'main', *self.connection) == saved, 'Сравнение изменило сессию')
        output = self.root / 'saved.xml'
        self.run('build', 'export', '--session', 'main', '--output', output, *self.connection)
        require(output.read_text('utf-8') == saved['build']['xml'], 'Экспорт XML не совпал')
        self.steps.append('immutable_compare_and_export')
        self.stop()
        backup = self.root / 'backup.db'
        self.run('storage', 'backup', '--profile', 'smoke', '--data-dir', self.root / 'data', '--file', backup)
        self.run('storage', 'verify', '--file', backup)
        self.run('storage', 'restore', '--profile', 'restored', '--data-dir', self.root / 'data', '--file', backup)
        self.start('restored')
        require(self.run(*calculate, *self.connection) == saved, 'Восстановление потеряло результат')
        self.stop()
        self.steps.append('backup_verify_restore_exact_retry')
        return {'version': version, 'steps': self.steps, 'total_dps': saved['build']['calculation']['stats']['TotalDPS'],
                'input_sha256': saved['build']['input_sha256']}


def smoke(archive, tag):
    with tempfile.TemporaryDirectory(prefix='pob2-release-') as directory:
        root = Path(directory).resolve()
        package_root = root / 'package'
        manifest, binary = unpack(archive, package_root, tag)
        goos = {'Windows': 'windows', 'Linux': 'linux', 'Darwin': 'darwin'}[platform.system()]
        arch = {'AMD64': 'amd64', 'x86_64': 'amd64', 'arm64': 'arm64', 'aarch64': 'arm64'}[platform.machine()]
        require((goos, arch) == (manifest['goos'], manifest['goarch']), 'Архив не соответствует настоящей платформе')
        work = root / 'work'
        work.mkdir()
        scenario = Scenario(binary, work)
        try:
            result = scenario.execute(package_root, tag)
        finally:
            scenario.close()
        return dict(result, tag=tag, commit=manifest['commit'], platform=platform.platform(), machine=platform.machine(),
                    archive_sha256=hashlib.sha256(archive.read_bytes()).hexdigest(),
                    checked_at=datetime.now(timezone.utc).isoformat(),
                    run_url=(os.environ['GITHUB_SERVER_URL'] + '/' + os.environ['GITHUB_REPOSITORY']
                             + '/actions/runs/' + os.environ['GITHUB_RUN_ID']) if os.environ.get('GITHUB_RUN_ID') else None)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--archive', required=True, type=Path)
    parser.add_argument('--tag', required=True)
    parser.add_argument('--report', required=True, type=Path)
    args = parser.parse_args()
    report = smoke(args.archive, args.tag)
    args.report.parent.mkdir(parents=True, exist_ok=True)
    args.report.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n', 'utf-8')
    print('Релизный сценарий пройден:', report['platform'])


if __name__ == '__main__':
    main()
