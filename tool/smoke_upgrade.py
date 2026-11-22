"""Нативная приёмка обновления PoB, отказа и отката с базой v0.0.5."""

import json
import sys
import tempfile
from pathlib import Path

from reference_env import CACHE, ROOT
from reference_upgrade import portable_archive, unpack_portable
from smoke_multigroup import Matrix
from smoke_release import Scenario, assert_lua_source, require, verify_compatibility_fixture

PROFILE = 'pob-update'
SESSION = 'main'
CATALOG = {
    'gems': ('Ice Nova', 'Spark', 'Fireball', 'Frostbolt', 'Concentrated Area', 'Rolling Slam',
             'Earthquake', 'Skeletal Arsonist', 'Bone Cage'),
    'items': ('Wooden Club', 'Smithing Hammer', 'Rusted Cuirass', 'Iron Cuirass'),
}


def pinned_releases(fixtures):
    directory = fixtures / 'testdata/reference/0.0.6'
    lock = json.loads((directory / 'releases.lock.json').read_text('utf-8'))
    previous_cases = json.loads((directory / 'previous.golden.json').read_text('utf-8'))['cases']
    return lock, previous_cases


def check_source(value, release):
    assert_lua_source(value, dict(tag=release['version'], commit=release['commit']))
    require(value['data_sha256'] == release['sha256'], 'Нет привязки расчёта к закреплённому пакету PoB')


def expect_installation(state, installed, running='', previous=''):
    require(state.get('installed_version', '') == installed, 'Неожиданная установленная версия PoB')
    require(state.get('running_version', '') == running, 'Неожиданная версия PoB работающего процесса')
    require(state.get('previous_version', '') == previous, 'Неожиданная предыдущая установка PoB')


def offline_environment(scenario):
    environment = {key: value for key, value in scenario.environment.items()
                   if key.upper() not in {'HTTP_PROXY', 'HTTPS_PROXY', 'NO_PROXY'}}
    return dict(environment, HTTP_PROXY='http://127.0.0.1:0', HTTPS_PROXY='http://127.0.0.1:0')


def update_target(release):
    return 'pob@' + release['version']


def install_previous(scenario, fixtures, release):
    state = scenario.run('update', '--target', update_target(release), timeout=365)['pob']
    require(state['status'] == 'updated' and state['installed_commit'] == release['commit'], 'Не установлен исходный PoB')
    expect_installation(state, release['version'])
    directory = fixtures / 'testdata/compat/0.0.5'
    original = verify_compatibility_fixture(directory, 'v0.0.5')
    scenario.run('storage', 'restore', '--profile', PROFILE, '--data-dir', scenario.root / 'data',
                 '--file', directory / 'sessions.db')
    scenario.start(PROFILE)
    require(scenario.snapshot(SESSION) == original, 'Открытие изменило исходную базу v0.0.5')
    require(scenario.calculate(SESSION, 2, 'calculate') == original, 'Потерян ответ запроса v0.0.5')
    return original


def update_live(scenario, previous, new):
    scenario.run('update', '--target', update_target(new), code=1, raw=True,
                 environment=offline_environment(scenario))
    local = scenario.run('doctor', environment=offline_environment(scenario))['pob']
    expect_installation(local, previous['version'])
    result = scenario.run('update', '--target', update_target(new), *scenario.connection, timeout=365)
    state = result['pob']
    require(state['status'] == 'updated' and state['restart_required'], 'PoB не обновлён для следующего запуска')
    expect_installation(state, new['version'], previous['version'], previous['version'])
    assert_restart_pending(scenario, previous, new)
    scenario.stop()


def assert_restart_pending(scenario, previous, new):
    report = scenario.run('doctor', *scenario.connection)
    expect_installation(report['pob'], new['version'], previous['version'], previous['version'])
    require(report['restart_required'], 'Doctor не обнаружил необходимость перезапуска')
    check = scenario.run('update', '--check', '--target', 'pob-rollback', *scenario.connection)['pob']
    require(check['status'] == 'rollback_available', 'После обновления отсутствует откат')
    expect_installation(check, new['version'], previous['version'], previous['version'])
    from_mcp, = scenario.mcp_tools(SESSION, [('update', dict(target='pob-rollback'))])
    require(from_mcp['pob'] == check, 'MCP и CLI/gRPC показывают разные установки для отката')


def current_launch(scenario, release):
    env = scenario.environment
    if sys.platform == 'win32':
        base = Path(env['LOCALAPPDATA'])
    elif sys.platform == 'darwin':
        base = Path(env['HOME']) / 'Library/Caches'
    else:
        base = Path(env['XDG_CACHE_HOME'])
    name = release['version'] + '-' + release['sha256'][:12]
    root = base / 'pob2-engine' / 'pob' / name
    for path in (root / 'Launch.lua', root / 'src/Launch.lua'):
        if path.is_file():
            return path
    raise FileNotFoundError('Не найден файл запуска установленного PoB в ' + str(root))


def rollback_offline(scenario, matrix, previous, new):
    scenario.environment = offline_environment(scenario)
    launch = current_launch(scenario, new)
    disabled = launch.with_suffix('.disabled')
    launch.rename(disabled)
    try:
        scenario.run('calculate', '--file', matrix.fixture('attacker'), code=1, raw=True)
        state = scenario.run('update', '--target', 'pob-rollback', timeout=50)['pob']
        require(state['status'] == 'rolled_back' and state['restart_required'], 'Не выполнен откат без сервера и сети')
        expect_installation(state, previous['version'])
    finally:
        disabled.rename(launch)
    scenario.start(PROFILE)


def verify_replayed(scenario, original, old_calculation, new_calculation):
    for expected, revision, request_id in ((original, 2, 'calculate'), (old_calculation, 3, 'previous-pob'),
                                           (new_calculation, 4, 'updated-pob')):
        require(scenario.calculate(SESSION, revision, request_id) == expected, 'Изменился исходный ответ ' + request_id)
    require(scenario.snapshot(SESSION) == new_calculation, 'Повтор прежнего запроса откатил ревизию сессии')


def check_catalog(scenario, release):
    found = {}
    trees = set()
    for kind, names in CATALOG.items():
        found[kind] = []
        for name in names:
            page = scenario.run('catalog', '--kind', kind, '--query', name, *scenario.connection)
            check_source(page, release)
            require(any(entry['name'] == name for entry in page['entries']), 'В каталоге не найдено: ' + name)
            trees.add(page['tree_version'])
            found[kind].append(name)
    require(len(trees) == 1, 'Одновременно использованы разные версии дерева')
    return dict(tree_version=trees.pop(), entries=found)


def run_upgrade_cycle(scenario, fixtures):
    lock, cases = pinned_releases(fixtures)
    previous, new = lock['previous'], lock['updated']
    old_matrix = Matrix(scenario, fixtures, lambda result: check_source(result, previous), cases)
    new_matrix = Matrix(scenario, fixtures, lambda result: check_source(result, new))
    original = install_previous(scenario, fixtures, previous)
    old = scenario.calculate(SESSION, original['revision'], 'previous-pob')
    old_matrix.check_snapshot(old, 'attacker')
    old_catalog = check_catalog(scenario, previous)
    update_live(scenario, previous, new)
    scenario.start(PROFILE)
    updated = scenario.calculate(SESSION, old['revision'], 'updated-pob')
    new_matrix.check_snapshot(updated, 'attacker')
    new_catalog = check_catalog(scenario, new)
    verify_replayed(scenario, original, old, updated)
    scenario.stop()
    rollback_offline(scenario, old_matrix, previous, new)
    verify_replayed(scenario, original, old, updated)
    restored = scenario.calculate(SESSION, updated['revision'], 'restored-pob')
    old_matrix.check_snapshot(restored, 'attacker')
    scenario.stop()
    return dict(previous=previous['version'], updated=new['version'], restored=previous['version'],
                source_release='v0.0.5', input_sha256=restored['build']['input_sha256'], end_revision=restored['revision'],
                offline_rollback=True, exact_retries=True, rejected_update_preserved_installation=True,
                catalogs={previous['version']: old_catalog, new['version']: new_catalog})


def check_upgrade(parent, fixtures):
    workspace = parent.root / 'pob-update'
    workspace.mkdir()
    scenario = Scenario(parent.binary, workspace)
    scenario.environment.pop('POB2_ENGINE_CALCULATOR')
    try:
        return run_upgrade_cycle(scenario, fixtures)
    finally:
        scenario.close()


def check_older_corpus(binary):
    lock, cases = pinned_releases(ROOT)
    release = lock['previous']
    archive = portable_archive(lock, release)
    with tempfile.TemporaryDirectory(prefix='pob-upgrade-check-', dir=CACHE) as directory:
        workspace = Path(directory)
        portable = workspace / 'portable'
        unpack_portable(archive, portable)
        scenario = Scenario(binary.resolve(), workspace)
        scenario.environment.pop('POB2_ENGINE_CALCULATOR')
        scenario.environment['POB2_ENGINE_POB_DIR'] = str(portable)
        native = CACHE / 'linux-x86_64/native'
        if sys.platform != 'win32' and native.is_dir():
            scenario.environment['POB2_ENGINE_POB_NATIVE_DIR'] = str(native)
        matrix = Matrix(scenario, ROOT, lambda result: require(result['data_id'] == 'pob-lua-local', 'Неверный источник'), cases)
        try:
            scenario.start(PROFILE)
            for name in cases:
                matrix.calculate(name, matrix.open(name, name), name)
            scenario.stop()
        finally:
            scenario.close()
    print('Движок сверил с эталоном PoB', release['version'], 'состояний:', len(cases))


if __name__ == '__main__':
    check_older_corpus(Path(sys.argv[1]))
