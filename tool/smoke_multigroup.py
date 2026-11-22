"""Проверить многогрупповые сборки через выпущенный CLI, MCP и gRPC."""

from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
from statistics import median
from time import perf_counter
import xml.etree.ElementTree as ET

from reference_lua import compare
from smoke_catalog import require

BASICS = ('caster', 'attacker', 'minion')


def issue_messages(result, code):
    return sorted(issue['message'] for issue in result['issues'] if issue['code'] == code)


def expected_exit(case):
    return 4 if case['warnings'] or case['unsupported'] or case['gem_errors'] else 0


def check_result(result, case, label):
    compare(case['skill'], result['skill'], label + '.skill')
    compare('partial' if expected_exit(case) else 'complete', result['status'], label + '.status')
    compare(case['stats'], {name: result['stats'].get(name) for name in case['stats']}, label + '.stats')
    for source, code in (('warnings', 'pob_warning'), ('unsupported', 'unsupported_modifier'),
                         ('gem_errors', 'pob_gem_error')):
        compare(sorted(case[source]), issue_messages(result, code), label + '.' + source)
    require(len(result['issues']) == sum(len(case[key]) for key in ('warnings', 'unsupported', 'gem_errors')),
            label + ': неожиданные сообщения диагностики')


def xml_structure(node):
    return (node.tag, node.attrib, (node.text or '').strip(), [xml_structure(child) for child in node])


class Matrix:
    def __init__(self, scenario, fixtures, check_source, cases=None):
        self.scenario = scenario
        self.directory = fixtures / 'testdata/reference/0.0.5'
        self.cases = cases if cases is not None else json.loads((self.directory / 'golden.json').read_text('utf-8'))['cases']
        self.check_source = check_source

    def fixture(self, name):
        path = self.directory / (name + '.xml')
        require(hashlib.sha256(path.read_bytes()).hexdigest() == self.cases[name]['input_sha256'],
                'Хеш эталонной сборки изменился: ' + name)
        return path

    def check_snapshot(self, snapshot, name):
        view, expected = snapshot['build'], self.cases[name]
        model = view['model']
        for name_in_model in ('main_group', 'active_sets', 'weapon_set'):
            compare(expected[name_in_model], model[name_in_model], name + '.' + name_in_model)
        compare(expected['enabled_groups'], [group['enabled'] for group in model['groups']], name + '.groups')
        saved = xml_structure(ET.fromstring(view['xml']))
        reference = xml_structure(ET.parse(self.fixture(name)).getroot())
        require(saved == reference, 'Сохранённый XML отличается от ожидаемого: ' + name)
        self.verify_result(view['calculation'], name)
        require(view['calculation']['input_sha256'] == view['input_sha256'] and
                view['calculated_from_revision'] == view['input_revision'], 'Потеряна привязка расчёта к XML')

    def verify_result(self, result, name):
        self.check_source(result)
        check_result(result, self.cases[name], name)

    def run_build(self, operation, session, *args, **options):
        return self.scenario.run('build', operation, '--session', session, *args, *self.scenario.connection, **options)

    def mcp(self, session, tool, **arguments):
        result, = self.scenario.mcp_tools(session, [(tool, arguments)])
        return result

    def open(self, session, name, path=None):
        path = self.fixture(name) if path is None else path
        self.scenario.run('session', 'create', '--session', session, '--name', session,
                          '--request-id', 'create-' + session, *self.scenario.connection)
        imported = self.run_build('import', session, '--revision', 1, '--file', path, '--request-id', 'import')
        from_mcp = self.mcp(session, 'import_build', expected_revision=1, request_id='import', xml=path.read_text('utf-8'))
        require(imported == from_mcp and imported['revision'] == 2, 'Импорт CLI/gRPC и MCP различается')
        return imported

    def edit(self, session, before, name, edit):
        revision = before['revision']
        edited = self.mcp(session, 'edit_build', expected_revision=revision, request_id=name, edit=edit)
        path = self.edit_file(name, edit)
        other = self.run_build('edit', session, '--revision', revision, '--file', path, '--request-id', name)
        require(edited == other and edited['revision'] == revision + 1, 'Правки CLI/gRPC и MCP различаются')
        return self.calculate(session, edited, name)

    def edit_file(self, name, edit):
        path = self.scenario.root / (name + '.json')
        path.write_text(json.dumps(edit, ensure_ascii=False), 'utf-8')
        return path

    def calculate(self, session, before, name):
        revision = before['revision']
        request = f'{name}-calculate-{revision}'
        saved = self.calculate_cli(session, revision, request, name)
        other = self.mcp(session, 'calculate_build', expected_revision=revision, request_id=request)
        require(saved == other and saved['revision'] == revision + 1, 'Расчёты CLI/gRPC и MCP различаются')
        self.check_snapshot(saved, name)
        return saved

    def calculate_cli(self, session, revision, request, name):
        return self.scenario.run('calculate', '--session', session, '--revision', revision, '--request-id', request,
                                 *self.scenario.connection, code=expected_exit(self.cases[name]))

    def preview(self, saved):
        support, session, revision = 'Concentrated Area', 'caster', saved['revision']
        result = self.scenario.run('preview-support', '--session', session, '--revision', revision,
                                   '--support', support, '--level', 1, '--quality', 20,
                                   *self.scenario.connection)
        other = self.mcp(session, 'preview_support', expected_revision=revision, support_id=support, level=1, quality=20)
        require(result == other and result['compatible'] and result['group_id'] == 1,
                'Предпросмотр CLI/gRPC и MCP различается или выбрана другая группа')
        self.check_comparison(session, saved, result, 'caster-preview', 'FullDPS')

    def compare_weapons(self, saved, edit):
        session, revision = 'attacker', saved['revision']
        path = self.edit_file('weapon-comparison', edit)
        result = self.run_build('compare', session, '--revision', revision, '--file', path, '--edit')
        other = self.mcp(session, 'compare_build', expected_revision=revision, edit=edit)
        require(result == other, 'Сравнения CLI/gRPC и MCP различаются')
        self.check_comparison(session, saved, result, 'attacker-swap', 'TotalDPS')

    def check_comparison(self, session, saved, response, name, stat):
        result = response['comparison']
        before = saved['build']['calculation']
        require(before == result['before'] and response['revision'] == saved['revision'], 'Подменён исходный снимок')
        self.verify_result(result['after'], name)
        difference = self.cases[name]['stats'][stat] - before['stats'][stat]
        delta = next(item for item in result['deltas'] if item['stat'] == stat)
        compare(difference, delta['difference'], name + '.delta.' + stat)
        require(self.scenario.snapshot(session) == saved, 'Сравнение или предпросмотр изменили сессию')

    def roundtrip(self, saved, name):
        path = self.scenario.root / 'export-roundtrip.xml'
        self.run_build('export', saved['session_id'], '--output', path)
        other = self.mcp(saved['session_id'], 'export_build')
        require(path.read_text('utf-8') == other['xml'] == saved['build']['xml'], 'Экспорт CLI/gRPC/MCP изменил XML')
        imported = self.open('roundtrip', name, path)
        return self.calculate('roundtrip', imported, name)

    def measure(self, repeats=5):
        imported = {name: self.open('measure-' + name, name) for name in BASICS}
        return {name: self.measure_case(name, imported[name]['revision'], repeats) for name in BASICS}

    def measure_case(self, name, revision, repeats):
        samples = []
        for index in range(repeats + 1):
            started = perf_counter()
            saved = self.calculate_cli('measure-' + name, revision + index, f'measure-{index}', name)
            samples.append((perf_counter() - started) * 1000)
            self.verify_result(saved['build']['calculation'], name)
        return dict(first_ms=samples[0], repeated_ms=samples[1:], repeated_median_ms=median(samples[1:]))


def exercise_matrix(matrix):
    caster = matrix.calculate('caster', matrix.open('caster', 'caster'), 'caster')
    attack_imported = matrix.open('attacker', 'attacker')
    with ThreadPoolExecutor(max_workers=2) as pool:
        preview = pool.submit(matrix.preview, caster)
        attack = pool.submit(matrix.calculate, 'attacker', attack_imported, 'attacker')
        preview.result()
        attacker = attack.result()
    swapped = dict(selection=dict(weapon_set=2))
    matrix.compare_weapons(attacker, swapped)
    caster = matrix.edit('caster', caster, 'caster-main', dict(selection=dict(main_group=2)))
    caster = matrix.edit('caster', caster, 'caster-alternate', dict(selection=dict(sets=dict(skills=2), main_group=1)))
    attacker = matrix.edit('attacker', attacker, 'attacker-swap', swapped)
    attacker = matrix.edit('attacker', attacker, 'attacker-outfit', dict(selection=dict(sets=dict(items=2))))
    armour = ET.parse(matrix.fixture('attacker-armour')).find("Items/Item[@id='5']").text
    attacker = matrix.edit('attacker', attacker, 'attacker-armour', dict(equipment=dict(slot='Body Armour', raw=armour)))
    minion = matrix.calculate('minion', matrix.open('minion', 'minion'), 'minion')
    minion = matrix.edit('minion', minion, 'minion-resistant',
                         dict(config=dict(name='enemyFireResist', value=dict(type='number', value='40'))))
    return dict(caster=caster, attacker=attacker, minion=minion, roundtrip=matrix.roundtrip(caster, 'caster-alternate'))


def check_restart(matrix, saved):
    scenario = matrix.scenario
    journal = scenario.root / 'data/pob2-engine/profiles/multigroup/sessions.db'
    verified = scenario.run('storage', 'verify', '--file', journal)
    require(verified['records'] == sum(item['revision'] for item in saved.values()), 'Обнаружены лишние записи журнала')
    previous = scenario.environment
    scenario.environment = {name: value for name, value in previous.items()
                            if name.upper() not in {'HTTP_PROXY', 'HTTPS_PROXY', 'NO_PROXY'}}
    scenario.environment.update(HTTP_PROXY='http://127.0.0.1:0', HTTPS_PROXY='http://127.0.0.1:0')
    try:
        scenario.start('multigroup')
        for name, snapshot in saved.items():
            require(scenario.snapshot(name) == snapshot, 'Новый запуск изменил сессию: ' + name)
        previous_caster = saved['caster']
        old_revision = previous_caster['revision'] - 1
        replay = matrix.calculate_cli('caster', old_revision, f'caster-alternate-calculate-{old_revision}', 'caster-alternate')
        require(replay == previous_caster, 'Потерян сохранённый ответ после перезапуска')
        updated = matrix.calculate('caster', previous_caster, 'caster-alternate')
        scenario.stop()
        return updated['revision']
    finally:
        scenario.close()
        scenario.environment = previous


def check_multigroup(scenario, fixtures, check_source):
    matrix = Matrix(scenario, fixtures, check_source)
    scenario.start('multigroup')
    saved = exercise_matrix(matrix)
    scenario.stop()
    restarted_revision = check_restart(matrix, saved)
    scenario.start('multigroup-measure')
    timings = matrix.measure()
    scenario.stop()
    return dict(cases=list(matrix.cases), restarted_revision=restarted_revision, timings=timings,
                parallel_preview_and_calculation=True, no_extra_journal_records=True)
