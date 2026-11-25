"""Проверить, что PoB II принял именно этот кандидат на всех платформах."""

import argparse
import json
import re
import subprocess
import time
from urllib.request import Request, urlopen


APP = 'Hecatoncheir/pob2_ui'
APP_WORKFLOW = '.github/workflows/ci.yml'
PLATFORMS = ('windows-x64', 'linux-x64', 'macos-arm64')
APP_STEPS = (
    'Загрузить комплектный движок',
    'Проверить общую XML-сборку с опубликованным движком',
    'Проверить фактическое завершение через MCP',
    'Проверить установку и восстановление сборки с Lua',
    'Предпросмотр и очистка профиля',
    'Сравнение путей дерева с настоящим Lua',
)


def git(*args):
    return subprocess.check_output(['git', *args])


def read_url(url):
    with urlopen(Request(url, headers={'User-Agent': 'pob2-engine-release'}), timeout=30) as response:
        return response.read()


def app_api(path):
    return json.loads(read_url(f'https://api.github.com/repos/{APP}/{path}'))


def require(condition, message):
    if not condition:
        raise SystemExit(message)


def stable_tag(candidate):
    match = re.fullmatch(r'(v(?:0|[1-9]\d*)(?:\.(?:0|[1-9]\d*)){2})-rc\.[1-9]\d*', candidate)
    require(match, 'Укажите кандидат в формате vX.Y.Z-rc.N')
    return match[1]


def check_engine(candidate):
    release = json.loads(subprocess.check_output(['gh', 'api', f'repos/Hecatoncheir/pob2_engine/releases/tags/{candidate}']))
    require(release['prerelease'] and not release['draft'], 'Кандидат должен быть опубликован')
    candidate_sha = git('rev-parse', f'{candidate}^{{}}').decode().strip()
    runs = json.loads(subprocess.check_output([
        'gh', 'run', 'list', '--workflow', 'release.yml', '--branch', candidate,
        '--json', 'headSha,conclusion', '--limit', '10',
    ]))
    require(any(run['headSha'] == candidate_sha and run['conclusion'] == 'success' for run in runs),
            'Проверка опубликованного кандидата должна завершиться успешно')
    changed = git('diff', '--name-only', candidate_sha, 'HEAD').decode().splitlines()
    notes = {'CHANGELOG.md', 'TODO.md', f'todo_{stable_tag(candidate)[1:]}.md'}
    require(set(changed) <= notes, 'Код стабильного выпуска должен совпадать с принятым кандидатом')


def check_app_run(run, candidate):
    head = run.get('head_repository') or {}
    require(head.get('full_name') == APP and run['path'] == APP_WORKFLOW,
            'Нужен прогон проверок из репозитория PoB II')
    require(run['event'] == 'workflow_dispatch' and run['display_title'] == f'Приёмка PoB II с {candidate}',
            'Нужна полная приёмка PoB II с указанным кандидатом и закреплёнными компонентами')


def wait_for_app(candidate, run_id):
    for _ in range(45):
        run = app_api(f'actions/runs/{run_id}')
        check_app_run(run, candidate)
        if run['status'] == 'completed':
            require(run['conclusion'] == 'success', 'Приёмка PoB II завершилась с ошибкой')
            return run
        time.sleep(120)
    raise SystemExit('Приёмка PoB II не завершилась за 90 минут')


def check_app_jobs(jobs):
    for platform in PLATFORMS:
        job = next((job for job in jobs if job['name'] == platform), {})
        require(job.get('conclusion') == 'success', f'Не пройдена приёмка {platform}')
        passed = {step['name'] for step in job['steps'] if step['conclusion'] == 'success'}
        require(set(APP_STEPS) <= passed, f'Неполная приёмка {platform}')


def check_app_api(candidate, revision):
    root = f'https://raw.githubusercontent.com/{APP}/{revision}'
    pinned = read_url(f'{root}/tool/engine-release.txt').decode('ascii').strip()
    require(pinned == candidate, 'В PoB II должен быть закреплён принятый кандидат')
    entries = app_api(f'git/trees/{revision}?recursive=1')['tree']
    prefix = 'packages/pob2_engine_api/lib/'
    app = {entry['path'].removeprefix(prefix): entry['sha'] for entry in entries
           if entry['type'] == 'blob' and entry['path'].startswith(prefix)}
    engine = 'gen/dart/lib/'
    local = {}
    for entry in git('ls-tree', '-r', 'HEAD', engine).decode().splitlines():
        metadata, path = entry.split('\t')
        local[path.removeprefix(engine)] = metadata.split()[2]
    require(local and local == app, 'Сгенерированный Dart API PoB II отличается от кандидата')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('candidate')
    parser.add_argument('app_run', type=int)
    args = parser.parse_args()
    tag = stable_tag(args.candidate)
    require(args.app_run > 0, 'Укажите номер прогона PoB II')
    check_engine(args.candidate)
    run = wait_for_app(args.candidate, args.app_run)
    jobs = app_api(f'actions/runs/{args.app_run}/jobs?filter=latest&per_page=100')
    check_app_jobs(jobs['jobs'])
    check_app_api(args.candidate, run['head_sha'])
    print(f'{tag}. PoB II прошёл приёмку: {run["html_url"]}')


if __name__ == '__main__':
    main()
