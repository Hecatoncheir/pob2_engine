"""Проверить релизный бинарник на целевой ОС без среды разработки."""

import argparse
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timezone
import hashlib
import json
import math
import os
from pathlib import Path
import platform
import queue
import shutil
from statistics import median
import subprocess
import tempfile
import threading
from time import perf_counter
import zipfile

DATA_ID = 'pob2-0.0.2-spell-hits-v1'
SPELLS = (
    {'id': 'fireball', 'name': 'Fireball', 'file': 'testdata/reference/fireball.xml', 'dps': 8.9166666666667},
    {'id': 'frostbolt', 'name': 'Frostbolt', 'file': 'testdata/reference/0.0.2/cases/frostbolt-minimal.xml', 'dps': 9.1},
    {'id': 'spark', 'name': 'Spark', 'file': 'testdata/reference/0.0.2/cases/spark-minimal.xml', 'dps': 8.5642857142857},
)
MCP_TOOLS = {'get_snapshot', 'rename_build', 'import_build', 'edit_build', 'calculate_build',
             'compare_build', 'export_build', 'get_catalog', 'preview_support', 'doctor', 'update',
             'pob_editor_status', 'pob_editor_ui', 'pob_editor_app', 'pob_editor_components',
             'pob_editor_library', 'pob_editor_history', 'pob_editor_notes', 'pob_editor_files',
             'pob_editor_progression', 'pob_editor_job', 'get_tree_geometry'}


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
    require(entries.get(archive.name) == hashlib.sha256(archive.read_bytes()).hexdigest(), 'Файл не совпал с SHA256SUMS')


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
    require(not result.get('isError'), f"MCP вернул ошибку инструмента: {result.get('content')}")
    return result


def initialize_mcp(send, messages):
    send({'id': 1, 'method': 'initialize', 'params': {'protocolVersion': '2025-11-25',
          'capabilities': {}, 'clientInfo': {'name': 'release-smoke', 'version': '2'}}})
    mcp_result(messages, 1)
    send({'method': 'notifications/initialized'})
    send({'id': 2, 'method': 'tools/list', 'params': {}})
    tools = mcp_result(messages, 2)['tools']
    require({tool['name'] for tool in tools} == MCP_TOOLS, 'Неверный набор инструментов MCP')
    send({'id': 3, 'method': 'tools/call', 'params': {'name': 'pob_editor_status', 'arguments': {}}})
    answer = receive(messages)
    result = answer.get('result', {})
    require(answer.get('id') == 3 and result.get('isError'), 'Без разрешённого окна команда MCP должна дать ошибку')


def check_calculation(result, dps):
    require(result['status'] == 'complete' and result['data_id'] == DATA_ID, 'Неверный статус или набор данных')
    require(math.isclose(result['stats']['TotalDPS'], dps, rel_tol=1e-10, abs_tol=1e-9), 'DPS не совпал с эталоном Lua')


def verify_compatibility_fixture(directory, release='v0.0.1'):
    manifest = json.loads((directory / 'manifest.json').read_text('utf-8'))
    require(manifest['release'] == release and manifest['schema'] == 1, 'Неверный эталон обновления')
    require(set(manifest['files']) == {'sessions.db', 'snapshot.json'}, 'Неполный эталон обновления')
    for name, digest in manifest['files'].items():
        require(hashlib.sha256((directory / name).read_bytes()).hexdigest() == digest, 'Изменён эталон: ' + name)
    return json.loads((directory / 'snapshot.json').read_text('utf-8'))


def check_updated_build(previous, updated):
    require(updated['revision'] == 4, 'Новая версия не сохранила расчёт')
    old, new = previous['build'], updated['build']
    require(new['xml'] == old['xml'] and new['input_sha256'] == old['input_sha256'], 'Обновление изменило входы')
    require(new['calculated_from_revision'] == old['input_revision'], 'Потеряна привязка обновлённого расчёта')
    return old['calculation'], new['calculation']


def check_updated_snapshot(previous, updated):
    old_result, new_result = check_updated_build(previous, updated)
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
                            if key.upper() not in {'PATH', 'HOME', 'USERPROFILE', 'APPDATA', 'LOCALAPPDATA',
                                                   'XDG_CONFIG_HOME', 'XDG_CACHE_HOME'}
                            and not key.upper().startswith(('POB', 'PYTHON'))}
        self.environment.update(PATH='', HOME=str(root), USERPROFILE=str(root), APPDATA=str(root),
                                LOCALAPPDATA=str(root), XDG_CONFIG_HOME=str(root / 'config'),
                                XDG_CACHE_HOME=str(root / 'cache'),
                                POB_DIR=str(root / 'missing-pob'), POB2_ENGINE_CALCULATOR='go')
        if token := os.environ.get('POB2_ENGINE_POB_GITHUB_TOKEN'):
            self.environment['POB2_ENGINE_POB_GITHUB_TOKEN'] = token
        self.server = None
        self.connection = []
        self.steps = []

    def run(self, *args, code=0, raw=False, timeout=20, environment=None):
        result = subprocess.run([str(self.binary), *map(str, args)], cwd=self.root,
                                env=self.environment if environment is None else environment,
                                capture_output=True, encoding='utf-8', timeout=timeout)
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
        doctor, edited = self.mcp_tools(session, [('doctor', {}), ('edit_build', damage_edit())])
        require(doctor['status'] == 'ok' and doctor['server']['storage_mode'] == 'disk', 'MCP doctor не проверил сервер')
        require(edited['revision'] == 3, 'MCP не изменил сборку')

    def mcp_tools(self, session, calls):
        with_session = ['--attach', session, *self.connection]
        process = subprocess.Popen([str(self.binary), 'mcp', *with_session],
                                   cwd=self.root, env=self.environment, stdin=subprocess.PIPE,
                                   stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, text=True, encoding='utf-8')
        messages = queue.Queue()
        threading.Thread(target=line_reader, args=(process.stdout, messages), daemon=True).start()

        def send(message):
            process.stdin.write(json.dumps(dict(jsonrpc='2.0', **message)) + '\n')
            process.stdin.flush()

        try:
            initialize_mcp(send, messages)
            answers = [call_mcp(send, messages, request_id, name, args)
                       for request_id, (name, args) in enumerate(calls, start=4)]
            process.stdin.close()
            require(process.wait(timeout=10) == 0, 'MCP не завершился по EOF')
            return answers
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
        self.check_atomic_xml_creation(package_root)
        upgrade = self.upgrade_session(package_root)
        reports = [dict(value['report'], standalone=direct[session]) for session, value in saved.items()]
        fireball = saved['fireball']['snapshot']['build']
        return {'schema': 2, 'version': version, 'steps': self.steps, 'spells': reports, 'upgrade': upgrade,
                'total_dps': fireball['calculation']['stats']['TotalDPS'], 'input_sha256': fireball['input_sha256']}

    def check_atomic_xml_creation(self, package_root):
        self.start('from-xml')
        create = ('session', 'create', '--session', 'atomic', '--name', 'Fireball',
                  '--request-id', 'create', '--file', package_root / SPELLS[0]['file'])
        rejected = self.root / 'invalid.xml'
        rejected.write_text('<Invalid/>', 'utf-8')
        self.run(*create[:-1], rejected, *self.connection, code=1, raw=True)
        initial = self.run(*create, *self.connection)
        require(initial['revision'] == initial['build']['input_revision'] == 1,
                'Первая ревизия не содержит исходный XML')
        self.run('session', 'rename', '--session', 'atomic', '--name', 'После', '--revision', 1,
                 '--request-id', 'rename', *self.connection)
        self.stop()
        self.start('from-xml')
        require(self.run(*create, *self.connection) == initial, 'Не восстановлен первый снимок XML')
        require(self.snapshot('atomic')['revision'] == 2, 'Повтор создания откатил сессию')
        self.stop()
        self.steps.append('atomic_xml_create_and_exact_retry_after_restart')

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

    def upgrade_session(self, package_root, release='v0.0.1', lock=None, profile='upgrade', verify=None):
        directory = package_root / 'testdata/compat' / release.removeprefix('v')
        previous = verify_compatibility_fixture(directory, release)
        self.run('storage', 'restore', '--profile', profile, '--data-dir', self.root / 'data', '--file', directory / 'sessions.db')
        self.start(profile)
        require(self.snapshot('main') == previous, f'Открытие изменило сессию {release}')
        require(self.calculate('main', 2, 'calculate') == previous, 'Потерян прежний request_id')
        updated = self.calculate('main', 3, 'release-upgrade')
        if verify is not None:
            verify(previous, updated)
        elif lock is None:
            check_updated_snapshot(previous, updated)
        else:
            check_lua_updated_snapshot(previous, updated, lock)
        self.check_upgrade_retries(previous, updated)
        self.stop()
        self.start(profile)
        self.check_upgrade_retries(previous, updated)
        self.stop()
        self.verify_upgrade_journal(profile)
        require(verify_compatibility_fixture(directory, release) == previous, 'Обновление изменило исходный эталон')
        self.steps.append(f"published_{release[1:].replace('.', '')}_session_upgrade_and_restart")
        return {'source_release': release, 'before_data_id': previous['build']['calculation']['data_id'],
                'after_data_id': updated['build']['calculation']['data_id'], 'original_revision': previous['revision'],
                'current_revision': updated['revision'], 'exact_old_retry': True, 'restart_verified': True}

    def check_upgrade_retries(self, previous, updated):
        require(self.calculate('main', 2, 'calculate') == previous, 'Новый расчёт заменил первоначальный ответ')
        require(self.calculate('main', 3, 'release-upgrade') == updated, 'Повтор нового расчёта изменился')
        require(self.snapshot('main') == updated, 'Повтор старого запроса откатил текущую сессию')

    def verify_upgrade_journal(self, profile='upgrade'):
        database = self.root / 'data/pob2-engine/profiles' / profile / 'sessions.db'
        verified = self.run('storage', 'verify', '--file', database)
        require(verified['schema'] == '2' and verified['records'] == 4, 'Обновление повторно записало команды')


def native_target():
    goos = {'Windows': 'windows', 'Linux': 'linux', 'Darwin': 'darwin'}[platform.system()]
    arch = {'AMD64': 'amd64', 'x86_64': 'amd64', 'arm64': 'arm64', 'aarch64': 'arm64'}[platform.machine()]
    return goos, arch


def call_mcp(send, messages, request_id, name, arguments):
    send({'id': request_id, 'method': 'tools/call', 'params': {'name': name, 'arguments': arguments}})
    return mcp_result(messages, request_id)['structuredContent']


def damage_edit():
    return {'expected_revision': 2, 'request_id': 'damage', 'edit': {'config': {'name': 'customMods',
            'value': {'type': 'string', 'value': '100% increased Spell Damage'}}}}


def run_scenario(binary, fixtures, work, tag):
    work.mkdir()
    scenario = Scenario(binary, work)
    try:
        return scenario.execute(fixtures, tag)
    finally:
        scenario.close()


def report_metadata(result, path, tag, sha_field):
    return dict(result, tag=tag, platform=platform.platform(), machine=platform.machine(),
                checked_at=datetime.now(timezone.utc).isoformat(),
                run_url=(os.environ['GITHUB_SERVER_URL'] + '/' + os.environ['GITHUB_REPOSITORY']
                         + '/actions/runs/' + os.environ['GITHUB_RUN_ID']) if os.environ.get('GITHUB_RUN_ID') else None,
                **{sha_field: hashlib.sha256(path.read_bytes()).hexdigest()})


def smoke(archive, tag):
    with tempfile.TemporaryDirectory(prefix='pob2-release-') as directory:
        root = Path(directory).resolve()
        package_root = root / 'package'
        manifest, binary = unpack(archive, package_root, tag)
        require(native_target() == (manifest['goos'], manifest['goarch']), 'Архив не соответствует настоящей платформе')
        result = run_scenario(binary, package_root, root / 'work', tag)
        return report_metadata(dict(result, commit=manifest['commit']), archive, tag, 'archive_sha256')


def smoke_binary(source, tag):
    goos, arch = native_target()
    extension = '.exe' if goos == 'windows' else ''
    require(source.name == f'pob2_engine_{tag}_{goos}_{arch}{extension}', 'Бинарник не соответствует тегу и платформе')
    with tempfile.TemporaryDirectory(prefix='pob2-release-') as directory:
        root = Path(directory).resolve()
        binary = Path(shutil.copyfile(source, root / f'pob2-engine{extension}'))
        binary.chmod(0o755)
        scenario = Scenario(binary, root)
        doctor = scenario.run('doctor')
        require(doctor['status'] == 'ok' and doctor['data_id'] == DATA_ID, 'Doctor: каталог не прошёл проверку')
        require((doctor['os'], doctor['arch'], doctor['version']) == (goos, arch, tag), 'Doctor: неверная сборка')
        require(doctor.get('commit', '') == os.environ.get('GITHUB_SHA', doctor.get('commit', '')), 'Doctor: неверный commit')
        licenses = scenario.run('doctor', '--licenses')
        require(all(licenses.get(name) for name in ('LICENSE', 'THIRD_PARTY_NOTICES.md', '.licenses/runtime/manifest.json')),
                'Бинарник не содержит лицензионные уведомления')
        result = run_scenario(binary, Path(__file__).resolve().parents[1], root / 'work', tag)
        lua = check_native_lua(scenario)
        return report_metadata(dict(result, commit=doctor.get('commit', ''), doctor=doctor, lua=lua),
                               source, tag, 'binary_sha256')


def check_native_lua(scenario):
    from smoke_catalog import check_full_catalog
    from smoke_multigroup import Matrix, check_multigroup
    from smoke_upgrade import check_upgrade
    scenario.environment.pop('POB2_ENGINE_CALCULATOR')
    fixtures = Path(__file__).resolve().parents[1]
    lock = json.loads((fixtures / 'testdata/reference/sources.lock.json').read_text('utf-8'))['pob']
    results = check_lua_install(scenario, fixtures, lock)
    verify_lua_offline(scenario, fixtures, lock)
    diagnostics = check_lua_diagnostics(scenario, fixtures, lock)
    doctor = scenario.run('doctor')
    require(doctor['pob']['installed_version'] == lock['tag'], 'Doctor: неожиданная версия установленного PoB')
    scenario.start('lua')
    try:
        sessions = {spell['name']: check_lua_session(scenario, fixtures, spell, lock) for spell in SPELLS}
        open_build_session(scenario, 'catalog', 'Полный каталог', fixtures / 'testdata/reference/fireball.xml')
        catalog = check_full_catalog(scenario, lock, assert_lua_result)
        selection = check_lua_selection(scenario, fixtures, lock)
        tree = check_lua_tree(scenario, fixtures)
        weapon_tree = check_lua_weapon_tree(scenario, fixtures)
        scenario.stop()
        upgrade = scenario.upgrade_session(fixtures, 'v0.0.2', lock)
        upgrade003 = scenario.upgrade_session(fixtures, 'v0.0.3', lock, 'upgrade003')
        check_source = lambda result: assert_lua_source(result, lock)
        matrix = Matrix(scenario, fixtures, check_source)
        upgrade004 = scenario.upgrade_session(fixtures, 'v0.0.4', profile='upgrade004',
                                              verify=lambda old, new: check_multigroup_upgrade(matrix, old, new))
        multigroup = check_multigroup(scenario, fixtures, check_source)
        upgrade005 = check_upgrade(scenario, fixtures)
        upgrade006 = scenario.upgrade_session(fixtures, 'v0.0.6', profile='upgrade006',
                                              verify=lambda old, new: check_same_lua_revision(old, new, lock))
        timings = {
            'lua': measure_binary_calculations(scenario, fixtures, 'luajit',
                                               lambda result, spell: assert_lua_result(result, spell['name'], spell['dps'], lock)),
            'go': measure_binary_calculations(scenario, fixtures, 'go',
                                              lambda result, spell: check_calculation(result, spell['dps'])),
        }
    finally:
        scenario.close()
    return dict(version=doctor['pob']['installed_version'], commit=lock['commit'],
                archive_sha256=selection['data_sha256'], dps=results, sessions=sessions, diagnostics=diagnostics,
                selected_set_dps=selection['stats']['TotalDPS'], upgrade=upgrade, upgrade003=upgrade003,
                upgrade004=upgrade004, upgrade005=upgrade005, upgrade006=upgrade006,
                multigroup=multigroup, catalog=catalog, tree=tree, weapon_tree=weapon_tree, timings=timings,
                parallel_first_start=True, reused_cache_offline=True)


def check_same_lua_revision(previous, updated, lock):
    old, new = check_updated_build(previous, updated)
    assert_lua_source(old, lock)
    assert_lua_source(new, lock)
    require(old['data_sha256'] == new['data_sha256'], 'При обновлении изменился архив PoB')
    require(old['issues'] == new['issues'], 'При обновлении изменилась диагностика PoB')


def check_multigroup_upgrade(matrix, previous, updated):
    check_updated_build(previous, updated)
    matrix.check_snapshot(previous, 'attacker')
    matrix.check_snapshot(updated, 'attacker')


def open_build_session(scenario, session, name, fixture):
    scenario.run('session', 'create', '--session', session, '--name', name, '--request-id', 'create-' + session,
                 *scenario.connection)
    scenario.run('build', 'import', '--session', session, '--revision', 1, '--request-id', 'import',
                 '--file', fixture, *scenario.connection)


def measure_binary_calculations(scenario, fixtures, engine, verify_result):
    scenario.environment['POB2_ENGINE_CALCULATOR'] = engine
    started = perf_counter()
    scenario.start('timing-' + engine)
    ready_ms = (perf_counter() - started) * 1000
    timings = {spell['id']: measure_session_calculations(scenario, fixtures, spell, verify_result) for spell in SPELLS}
    scenario.stop()
    return dict(server_ready_ms=ready_ms, spells=timings)


def measure_session_calculations(scenario, fixtures, spell, verify_result):
    session = spell['id']
    open_build_session(scenario, session, spell['name'], fixtures / spell['file'])
    samples = []
    for revision in range(2, 8):
        started = perf_counter()
        snapshot = scenario.calculate(session, revision, f'timing-{revision}')
        samples.append((perf_counter() - started) * 1000)
        require(snapshot['revision'] == revision + 1, 'Замер получил повтор старого расчёта')
        verify_result(snapshot['build']['calculation'], spell)
    first, *warm = samples
    return dict(first_calculate_ms=first, repeated_calculate_ms=warm, repeated_median_ms=median(warm))


def check_lua_session(scenario, fixtures, spell, lock):
    session = spell['id']
    open_build_session(scenario, session, spell['name'], fixtures / spell['file'])
    calls = [('edit_build', damage_edit()),
             ('calculate_build', {'expected_revision': 3, 'request_id': 'calculate'}),
             ('get_snapshot', {})]
    edited, calculated, snapshot = scenario.mcp_tools(session, calls)
    require(edited['revision'] == 3 and calculated['revision'] == 4, 'MCP не сохранил правку и расчёт')
    repeated = scenario.calculate(session, 3, 'calculate')
    require(calculated == snapshot == repeated, 'CLI и MCP получили разные снимки расчёта gRPC')
    dps = assert_lua_result(calculated['build']['calculation'], spell['name'], spell['dps'] * 2, lock)
    preview = scenario.preview_support(session, repeated, 'Rapid Casting I')
    return dict(edited_dps=dps, preview=preview, cli_replayed_mcp_calculation=True)


def check_lua_selection(scenario, fixtures, lock):
    fixture = fixtures / 'testdata/reference/0.0.2/groups/cases/druid-group-1.xml'
    open_build_session(scenario, 'sets', 'Druid', fixture)
    edit = {'selection': {'main_group': 1, 'sets': {'skills': 2, 'config': 2, 'items': 1, 'tree': 1}}}
    args = {'expected_revision': 2, 'request_id': 'sets', 'edit': edit}
    edited, = scenario.mcp_tools('sets', [('edit_build', args)])
    active = edited['build']['model']['active_sets']
    require(edited['revision'] == 3 and active['skills'] == active['config'] == 2,
            'После правки сервер не вернул выбранные наборы')
    calculated = scenario.calculate('sets', 3, 'calculate')['build']['calculation']
    assert_lua_result(calculated, 'Spark', 17.1285714285714, lock)
    return calculated


def check_lua_tree(scenario, fixtures):
    fixture = fixtures / 'testdata/reference/0.0.10/warrior.xml'
    open_build_session(scenario, 'tree', 'Warrior', fixture)
    baseline = scenario.snapshot('tree')
    allocated, preview = commit_tree_edit(scenario, baseline, 13397, 'allocate', {'added': [3936, 13397]})
    refunded, _ = commit_tree_edit(scenario, allocated, 3936, 'refund', {'removed': [3936, 13397]})
    require(refunded['build']['input_sha256'] == baseline['build']['input_sha256'],
            'Снятие ветви не восстановило исходный XML')
    before, after = preview['comparison']['before']['stats'], preview['comparison']['after']['stats']
    require(after['Str'] == before['Str'] + 5 and after['TotalDPS'] > before['TotalDPS'],
            'PoB не учёл силу и урон после назначения пассивов')
    return dict(added=[3936, 13397], removed=[3936, 13397], strength_gain=5,
                dps=[before['TotalDPS'], after['TotalDPS']], restored_xml=True,
                attributes=check_tree_attributes(scenario, refunded))


def check_tree_attributes(scenario, baseline):
    path = [1207, 3936, 13397]
    choices = [dict(node_id=node_id, after='dexterity') for node_id in (1207, 13397)]
    allocated, preview = commit_tree_edit(scenario, baseline, 1207, 'allocate',
                                         dict(added=path, attributes=choices), 'dexterity')
    verify_attribute_stats(preview, {'Dex': 10})
    changes = [dict(node_id=13397, before='dexterity', after='intelligence')]
    switched, preview = commit_tree_edit(scenario, allocated, 13397, 'set_attribute',
                                         dict(attributes=changes), 'intelligence')
    verify_attribute_stats(preview, {'Dex': -5, 'Int': 5})
    before = allocated['build']['model']
    after = switched['build']['model']
    require(after['nodes'] == before['nodes'], 'Смена атрибута изменила назначенные узлы')
    require(after['passive_attributes'].get('1207') == 'dexterity'
            and after['passive_attributes'].get('13397') == 'intelligence',
            'Состояние атрибутов не совпало с предпросмотром')
    return dict(allocated=choices, switched=changes, dexterity_gain=10, intelligence_gain=5)


def verify_attribute_stats(preview, expected):
    comparison = preview['comparison']
    before, after = comparison['before']['stats'], comparison['after']['stats']
    for stat in ('Str', 'Dex', 'Int'):
        require(after[stat] - before[stat] == expected.get(stat, 0),
                f'PoB вернул неожиданное изменение {stat}')


def check_lua_weapon_tree(scenario, fixtures):
    session = 'tree-weapons'
    fixture = fixtures / 'testdata/reference/0.0.12/witch-weapon-paths.xml'
    open_build_session(scenario, session, 'Witch', fixture)
    initial = scenario.snapshot(session)
    selected, = scenario.mcp_tools(session, [('edit_build', dict(
        expected_revision=initial['revision'], request_id='activate-weapon-two',
        edit={'selection': {'weapon_set': 2}},
    ))])
    mode = 'weapon_set_2'
    path = [1755, 18845, 41965, 51184]
    before_points = dict(passives=1, weapon_set_1=1, weapon_set_2=1)
    after_points = dict(passives=5, weapon_set_1=1, weapon_set_2=5)
    allocated, preview = commit_tree_edit(
        scenario, selected, 51184, 'allocate',
        dict(added=path, allocations=[dict(node_id=node_id, after=mode) for node_id in path],
             points=dict(before=before_points, after=after_points)),
        allocation_mode=mode, session=session, compare_exit=4,
    )
    restored, _ = commit_tree_edit(
        scenario, allocated, 18845, 'refund',
        dict(removed=path, allocations=[dict(node_id=node_id, before=mode) for node_id in path],
             points=dict(before=after_points, after=before_points)),
        allocation_mode=mode, session=session, compare_exit=4,
    )
    before, after = preview['comparison']['before'], preview['comparison']['after']
    require(abs(before['stats']['TotalDPS'] - 19.482916666667) < 1e-5 and abs(after['stats']['TotalDPS'] - 24.61) < 1e-5,
            'PoB не учёл ветвь второго комплекта оружия')
    require(restored['build']['input_sha256'] == selected['build']['input_sha256'],
            'Снятие ветви второго комплекта не восстановило исходный XML')
    return dict(mode=mode, active_weapon=2, path=path, points=preview['passive_change']['points'],
                dps=[before['stats']['TotalDPS'], after['stats']['TotalDPS']], restored_xml=True)


def commit_tree_edit(scenario, snapshot, node_id, action, change, attribute=None, *,
                     allocation_mode=None, session='tree', compare_exit=0):
    tree = dict(node_id=node_id, action=action)
    if attribute is not None:
        tree['attribute'] = attribute
    if allocation_mode is not None:
        tree['allocation_mode'] = allocation_mode
    edit = dict(tree_node=tree)
    revision = snapshot['revision']
    path = scenario.root / 'tree-edit.json'
    path.write_text(json.dumps(edit), encoding='utf-8')
    preview = scenario.run('build', 'compare', *scenario.connection, '--session', session,
                           '--revision', revision, '--edit', '--file', path, code=compare_exit)
    arguments = dict(expected_revision=revision, edit=edit)
    mcp, = scenario.mcp_tools(session, [('compare_build', arguments)])
    require(preview == mcp and preview['passive_change'] == change, 'CLI/gRPC и MCP должны показать ожидаемый путь')
    require(scenario.snapshot(session) == snapshot, 'Предпросмотр изменил сессию')
    applied, = scenario.mcp_tools(session, [('edit_build', dict(arguments, request_id=f'{action}-{node_id}'))])
    require(applied['revision'] == revision + 1 and
            applied['build']['input_sha256'] == preview['comparison']['after']['input_sha256'],
            'Правка дерева разошлась с предпросмотром')
    return applied, preview


def check_lua_install(scenario, fixtures, lock):
    with ThreadPoolExecutor(max_workers=2) as pool:
        first = [pool.submit(calculate_lua_file, scenario, fixtures, spell) for spell in SPELLS[:2]]
        results = [future.result() for future in first]
    results.append(calculate_lua_file(scenario, fixtures, SPELLS[2]))
    return {spell['name']: assert_lua_result(result, spell['name'], spell['dps'], lock)
            for spell, result in zip(SPELLS, results, strict=True)}


def calculate_lua_file(scenario, fixtures, spell, environment=None):
    return scenario.run('calculate', '--file', fixtures / spell['file'], '--timeout', '40s', timeout=355,
                        environment=environment)


def verify_lua_offline(scenario, fixtures, lock):
    environment = {key: value for key, value in scenario.environment.items()
                   if key.upper() not in {'HTTP_PROXY', 'HTTPS_PROXY', 'NO_PROXY'}}
    environment.update(HTTP_PROXY='http://127.0.0.1:0', HTTPS_PROXY='http://127.0.0.1:0')
    spell = SPELLS[0]
    calculated = calculate_lua_file(scenario, fixtures, spell, environment)
    assert_lua_result(calculated, spell['name'], spell['dps'], lock)
    checked = scenario.run('doctor', environment=environment)['pob']
    require(checked['status'] == 'unavailable' and checked['installed_version'] == lock['tag'],
            'Doctor должен показать установленный PoB при недоступности GitHub')
    update = scenario.run('update', '--check', environment=environment)
    require(update['status'] == 'partial' and update['engine']['status'] == 'unavailable'
            and update['pob']['installed_version'] == lock['tag'],
            'Update должен показать статус двигателя и установленную версию PoB без GitHub')


def assert_lua_result(result, name, dps, lock):
    assert_lua_source(result, lock)
    require(result['status'] == 'complete' and result['skill'] == name, f'Неполный расчёт PoB для {name}')
    actual = result['stats']['TotalDPS']
    require(math.isclose(actual, dps, abs_tol=1e-8), f'DPS {name}: {actual}')
    return actual


def assert_lua_source(result, lock):
    require(result['data_id'] == f"pob-lua-{lock['tag']}", 'Расчёт выполнен с неожиданной версией PoB')
    require(result['pob_commit'] == lock['commit'], 'Неожиданная ревизия PoB')
    require(len(result['data_sha256']) == 64, 'Нет SHA-256 архива PoB')


def check_lua_updated_snapshot(previous, updated, lock):
    old, new = check_updated_build(previous, updated)
    require(old['data_id'] == DATA_ID, 'Прежний расчёт выполнен с неожиданным набором данных')
    require(new['data_sha256'] != old['data_sha256'], 'Новый расчёт сохранил контрольную сумму прежнего каталога')
    assert_lua_result(new, 'Fireball', old['stats']['TotalDPS'], lock)


def check_lua_diagnostics(scenario, fixtures, lock):
    examples = {'gem-requirement-warning': 'pob_warning', 'unparsed-custom-modifier': 'unsupported_modifier'}
    for name, code in examples.items():
        file = fixtures / 'testdata/reference/cases' / (name + '.xml')
        result = scenario.run('calculate', '--file', file, '--timeout', '40s', code=4, timeout=45)
        assert_lua_source(result, lock)
        require(result['status'] == 'partial' and any(issue['code'] == code for issue in result['issues']),
                f'PoB не передал диагностику {code}')
    return examples


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument('--binary', type=Path)
    source.add_argument('--archive', type=Path)
    parser.add_argument('--tag', required=True)
    parser.add_argument('--report', required=True, type=Path)
    parser.add_argument('--checksums', type=Path, help='SHA256SUMS скачанного выпуска')
    args = parser.parse_args()
    if args.checksums:
        verify_checksum(args.binary or args.archive, args.checksums)
    report = smoke_binary(args.binary, args.tag) if args.binary else smoke(args.archive, args.tag)
    report['published_checksum_verified'] = args.checksums is not None
    args.report.parent.mkdir(parents=True, exist_ok=True)
    args.report.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n', encoding='utf-8', newline='\n')
    print('Релизный сценарий пройден:', report['platform'])


if __name__ == '__main__':
    main()
