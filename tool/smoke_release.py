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

DATA_ID = 'pob2-0.0.2-spell-hits-v1'
SPELLS = (
    {'id': 'fireball', 'name': 'Fireball', 'file': 'testdata/reference/fireball.xml', 'dps': 8.9166666666667},
    {'id': 'frostbolt', 'name': 'Frostbolt', 'file': 'testdata/reference/0.0.2/cases/frostbolt-minimal.xml', 'dps': 9.1},
    {'id': 'spark', 'name': 'Spark', 'file': 'testdata/reference/0.0.2/cases/spark-minimal.xml', 'dps': 8.5642857142857},
)
MCP_TOOLS = {'get_snapshot', 'rename_build', 'import_build', 'edit_build', 'calculate_build',
             'compare_build', 'export_build', 'get_catalog', 'preview_support'}


def require(condition, message):
    if not condition:
        raise ValueError(message)


def verify_checksum(archive, checksums):
    entries = {}
    for line in checksums.read_text('ascii').splitlines():
        digest, name = line.split(maxsplit=1)
        name = name.lstrip(' *')
        require(name not in entries, 'Повтор файла в SHA256SUMS')
        entries[name] = digest
    require(entries.get(archive.name) == hashlib.sha256(archive.read_bytes()).hexdigest(), 'Архив не совпал с SHA256SUMS')


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
        require(source.read('COMMIT.txt').decode('ascii').strip() == manifest['commit'], 'COMMIT.txt не совпал с манифестом')
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


def mcp_result(messages, request_id):
    response = receive(messages)
    require(response.get('id') == request_id and 'result' in response, 'Нет ответа MCP')
    result = response['result']
    require(not result.get('isError'), 'MCP вернул ошибку инструмента')
    return result


def initialize_mcp(send, messages):
    send({'id': 1, 'method': 'initialize', 'params': {'protocolVersion': '2025-11-25',
          'capabilities': {}, 'clientInfo': {'name': 'release-smoke', 'version': '2'}}})
    mcp_result(messages, 1)
    send({'method': 'notifications/initialized'})
    send({'id': 2, 'method': 'tools/list', 'params': {}})
    tools = mcp_result(messages, 2)['tools']
    require({tool['name'] for tool in tools} == MCP_TOOLS, 'Неверный набор девяти инструментов MCP')


def check_calculation(result, dps):
    require(result['status'] == 'complete' and result['data_id'] == DATA_ID, 'Неверный статус или набор данных')
    require(math.isclose(result['stats']['TotalDPS'], dps, rel_tol=1e-10, abs_tol=1e-9), 'DPS не совпал с эталоном Lua')


def verify_compatibility_fixture(directory):
    manifest = json.loads((directory / 'manifest.json').read_text('utf-8'))
    require(manifest['release'] == 'v0.0.1' and manifest['schema'] == 1, 'Неверный эталон обновления')
    require(set(manifest['files']) == {'sessions.db', 'snapshot.json'}, 'Неполный эталон обновления')
    for name, digest in manifest['files'].items():
        require(hashlib.sha256((directory / name).read_bytes()).hexdigest() == digest, 'Изменён эталон: ' + name)
    return json.loads((directory / 'snapshot.json').read_text('utf-8'))


def check_updated_snapshot(previous, updated):
    require(updated['revision'] == 4, 'Новая версия не сохранила расчёт')
    old, new = previous['build'], updated['build']
    require(new['xml'] == old['xml'] and new['input_sha256'] == old['input_sha256'], 'Обновление изменило входы')
    require(new['calculated_from_revision'] == old['input_revision'], 'Потеряна привязка обновлённого расчёта')
    old_result, new_result = old['calculation'], new['calculation']
    require(old_result['data_id'] == 'pob2-0.0.1-fireball-v1' and new_result['data_id'] == DATA_ID, 'Набор данных не обновился')
    require(new_result['data_sha256'] != old_result['data_sha256'], 'Не изменился хеш набора')
    check_calculation(new_result, old_result['stats']['TotalDPS'])
    require(new_result['stats'].keys() == old_result['stats'].keys(), 'Изменились показатели Fireball')
    for name, value in old_result['stats'].items():
        require(math.isclose(new_result['stats'][name], value, rel_tol=1e-10, abs_tol=1e-9), 'Изменился показатель: ' + name)


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

    def mcp_edit(self, session):
        process = subprocess.Popen([str(self.binary), 'mcp', '--attach', session, *self.connection],
                                   cwd=self.root, env=self.environment, stdin=subprocess.PIPE,
                                   stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, text=True, encoding='utf-8')
        messages = queue.Queue()
        threading.Thread(target=line_reader, args=(process.stdout, messages), daemon=True).start()

        def send(message):
            process.stdin.write(json.dumps(dict(jsonrpc='2.0', **message)) + '\n')
            process.stdin.flush()

        try:
            initialize_mcp(send, messages)
            send({'id': 3, 'method': 'tools/call', 'params': {'name': 'edit_build', 'arguments': {
                'expected_revision': 2, 'request_id': 'damage', 'edit': {'config': {'name': 'customMods',
                'value': {'type': 'string', 'value': '100% increased Spell Damage'}}}}}})
            result = mcp_result(messages, 3)
            require(result['structuredContent']['revision'] == 3, 'MCP не изменил сборку')
            process.stdin.close()
            require(process.wait(timeout=10) == 0, 'MCP не завершился по EOF')
        finally:
            if process.poll() is None:
                process.kill()
                process.wait(timeout=10)

    def execute(self, package_root, tag):
        version = self.run('--version')
        require(version['version'] == tag and version['protocol_version'] == 'pob2.v1', 'Версия бинарника не соответствует тегу')
        direct = self.standalone_spells(package_root)
        self.start('smoke')
        saved = {spell['id']: self.session_spell(package_root, spell) for spell in SPELLS}
        self.steps.append('three_spells_cli_mcp_grpc_shared_calculation')
        self.stop()
        self.restore_sessions(saved)
        upgrade = self.upgrade_session(package_root)
        reports = [dict(value['report'], standalone=direct[session]) for session, value in saved.items()]
        fireball = saved['fireball']['snapshot']['build']
        return {'schema': 2, 'version': version, 'steps': self.steps, 'spells': reports, 'upgrade': upgrade,
                'total_dps': fireball['calculation']['stats']['TotalDPS'], 'input_sha256': fireball['input_sha256']}

    def standalone_spells(self, package_root):
        results = {}
        for spell in SPELLS:
            result = self.run('calculate', '--file', package_root / spell['file'])
            check_calculation(result, spell['dps'])
            require(result['skill'] == spell['name'], 'Одноразовый расчёт подменил умение')
            results[spell['id']] = {'total_dps': result['stats']['TotalDPS'], 'input_sha256': result['input_sha256'],
                                    'data_id': result['data_id']}
        self.steps.append('three_spells_standalone_without_runtime')
        return results

    def calculate(self, session, revision, request):
        return self.run('calculate', '--session', session, '--revision', revision, '--request-id', request, *self.connection)

    def snapshot(self, session):
        return self.run('session', 'get', '--session', session, *self.connection)

    def session_spell(self, package_root, spell):
        session = spell['id']
        self.run('session', 'create', '--session', session, '--name', spell['name'], '--request-id', 'create-' + session, *self.connection)
        self.run('build', 'import', '--session', session, '--revision', 1, '--request-id', 'import',
                 '--file', package_root / spell['file'], *self.connection)
        self.mcp_edit(session)
        saved = self.calculate(session, 3, 'calculate')
        require(saved['revision'] == 4 and saved['build']['calculation']['status'] == 'complete', 'Расчёт не сохранён')
        check_calculation(saved['build']['calculation'], spell['dps'] * 2)
        catalog = self.run('catalog', '--kind', 'gems', '--query', spell['name'], *self.connection)
        require(catalog['data_id'] == DATA_ID and catalog['total'] == 1, 'Неверный каталог умения')
        previews = [self.preview_support(session, saved, support) for support in ('Rapid Casting I', 'Controlled Destruction')]
        compared = self.compare_export(spell, saved)
        return {'snapshot': saved, 'report': {'session_id': session, 'skill': spell['name'], 'revision': saved['revision'],
                'total_dps': saved['build']['calculation']['stats']['TotalDPS'], 'input_sha256': saved['build']['input_sha256'],
                'data_id': catalog['data_id'], 'previews': previews, 'comparison_after_dps': compared}}

    def preview_support(self, session, saved, support):
        result = self.run('preview-support', '--session', session, '--revision', 4, '--support', support, *self.connection)
        require(result['status'] == 'complete' and result['compatible'] and result['revision'] == 4, 'Предпросмотр поддержки не выполнен')
        before, after = result['comparison']['before'], result['comparison']['after']
        require(before['stats'] == saved['build']['calculation']['stats'], 'Предпросмотр подменил исходный расчёт')
        require(before['input_sha256'] == saved['build']['input_sha256'] and after['input_sha256'] != before['input_sha256'], 'Нет привязки предпросмотра')
        require(after['stats']['TotalDPS'] > 0, 'Предпросмотр не вернул показатели')
        if support == 'Rapid Casting I':
            require(after['stats']['Speed'] > before['stats']['Speed'], 'Поддержка не изменила скорость')
        else:
            require(after['stats']['CritChance'] == 0, 'Controlled Destruction не отключил критические попадания')
        require(self.snapshot(session) == saved, 'Предпросмотр изменил сессию')
        return {'support': support, 'before_dps': before['stats']['TotalDPS'], 'after_dps': after['stats']['TotalDPS'],
                'proposal_sha256': after['input_sha256']}

    def compare_export(self, spell, saved):
        session = spell['id']
        edit = self.root / (session + '-edit.json')
        edit.write_text('{"config":{"name":"customMods","value":{"type":"string","value":"200% increased Spell Damage"}}}', 'utf-8')
        compared = self.run('build', 'compare', '--session', session, '--revision', 4, '--edit', '--file', edit, *self.connection)
        check_calculation(compared['comparison']['before'], spell['dps'] * 2)
        check_calculation(compared['comparison']['after'], spell['dps'] * 3)
        require(self.snapshot(session) == saved, 'Сравнение изменило сессию')
        output = self.root / (session + '-saved.xml')
        self.run('build', 'export', '--session', session, '--output', output, *self.connection)
        require(output.read_text('utf-8') == saved['build']['xml'], 'Экспорт XML не совпал')
        return compared['comparison']['after']['stats']['TotalDPS']

    def restore_sessions(self, saved):
        backup = self.root / 'backup.db'
        self.run('storage', 'backup', '--profile', 'smoke', '--data-dir', self.root / 'data', '--file', backup)
        verified = self.run('storage', 'verify', '--file', backup)
        require(verified['schema'] == '2' and verified['records'] == 12, 'Неверный журнал трёх сессий')
        self.run('storage', 'restore', '--profile', 'restored', '--data-dir', self.root / 'data', '--file', backup)
        self.start('restored')
        for session, value in saved.items():
            require(self.snapshot(session) == value['snapshot'], 'Восстановление потеряло сессию')
            require(self.calculate(session, 3, 'calculate') == value['snapshot'], 'Восстановление потеряло результат запроса')
            value['report']['restored_exact_retry'] = True
        self.stop()
        self.steps.append('three_spells_catalog_preview_compare_export')
        self.steps.append('three_sessions_backup_verify_restore_exact_retry')

    def upgrade_session(self, package_root):
        directory = package_root / 'testdata/compat/0.0.1'
        previous = verify_compatibility_fixture(directory)
        self.run('storage', 'restore', '--profile', 'upgrade', '--data-dir', self.root / 'data', '--file', directory / 'sessions.db')
        self.start('upgrade')
        require(self.snapshot('main') == previous, 'Открытие изменило сессию выпущенного 0.0.1')
        require(self.calculate('main', 2, 'calculate') == previous, 'Потерян прежний request_id')
        updated = self.calculate('main', 3, 'release-upgrade')
        check_updated_snapshot(previous, updated)
        self.check_upgrade_retries(previous, updated)
        self.stop()
        self.start('upgrade')
        self.check_upgrade_retries(previous, updated)
        self.stop()
        self.verify_upgrade_journal()
        require(verify_compatibility_fixture(directory) == previous, 'Обновление изменило исходный эталон')
        self.steps.append('published_001_session_upgrade_and_restart')
        return {'source_release': 'v0.0.1', 'before_data_id': previous['build']['calculation']['data_id'],
                'after_data_id': updated['build']['calculation']['data_id'], 'original_revision': previous['revision'],
                'current_revision': updated['revision'], 'exact_old_retry': True, 'restart_verified': True}

    def check_upgrade_retries(self, previous, updated):
        require(self.calculate('main', 2, 'calculate') == previous, 'Новый расчёт заменил первоначальный ответ')
        require(self.calculate('main', 3, 'release-upgrade') == updated, 'Повтор нового расчёта изменился')
        require(self.snapshot('main') == updated, 'Повтор старого запроса откатил текущую сессию')

    def verify_upgrade_journal(self):
        database = self.root / 'data/pob2-engine/profiles/upgrade/sessions.db'
        verified = self.run('storage', 'verify', '--file', database)
        require(verified['schema'] == '2' and verified['records'] == 4, 'Обновление повторно записало команды')


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
    parser.add_argument('--checksums', type=Path, help='SHA256SUMS скачанного выпуска')
    args = parser.parse_args()
    if args.checksums:
        verify_checksum(args.archive, args.checksums)
    report = smoke(args.archive, args.tag)
    report['published_checksum_verified'] = args.checksums is not None
    args.report.parent.mkdir(parents=True, exist_ok=True)
    args.report.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n', encoding='utf-8', newline='\n')
    print('Релизный сценарий пройден:', report['platform'])


if __name__ == '__main__':
    main()
