"""Воспроизвести многогрупповые расчёты на предыдущем официальном PoB."""

import hashlib
from pathlib import Path, PurePosixPath
import shutil
import sys
import tempfile
import urllib.request
import zipfile

from reference_env import CACHE, DATA, read_json, verify
from reference_lua import compare, encode
from reference_multigroup import collect_cases

DIRECTORY = DATA / '0.0.6'


def portable_archive(lock, release):
    CACHE.mkdir(exist_ok=True)
    name = release['version'].removeprefix('v')
    path = CACHE / f'pob-{name}-portable.zip'
    if path.is_file():
        verify_archive(path, release)
        return path
    url = f"https://github.com/{lock['repository']}/releases/download/{release['version']}/{lock['asset']}"
    with tempfile.TemporaryDirectory(prefix='pob-download-', dir=CACHE) as directory:
        staged = Path(directory) / path.name
        with urllib.request.urlopen(url, timeout=120) as response, staged.open('wb') as output:
            shutil.copyfileobj(response, output)
        verify_archive(staged, release)
        staged.replace(path)
    return path


def verify_archive(path, release):
    if path.stat().st_size != release['size']:
        raise ValueError('Неверный размер пакета ' + release['version'])
    with path.open('rb') as source:
        digest = hashlib.file_digest(source, 'sha256').hexdigest()
    if digest != release['sha256']:
        raise ValueError('Неверный SHA-256 пакета ' + release['version'])


def unpack_portable(path, destination):
    with zipfile.ZipFile(path) as archive:
        for name in archive.namelist():
            entry = PurePosixPath(name)
            if entry.is_absolute() or '..' in entry.parts or ':' in name or '\\' in name:
                raise ValueError('Недопустимый путь в официальном архиве PoB: ' + name)
        archive.extractall(destination)


def execute(record):
    state = verify()
    lock = read_json(DIRECTORY / 'releases.lock.json')
    if state['pob_commit'] != lock['updated']['commit']:
        raise ValueError('Новый выпуск PoB не совпадает с основным эталоном')
    path = portable_archive(lock, lock['previous'])
    with tempfile.TemporaryDirectory(prefix='pob-previous-reference-', dir=CACHE) as directory:
        pob_dir = Path(directory)
        unpack_portable(path, pob_dir)
        cases = collect_cases(pob_dir)
    actual = dict(schema=1, pob_commit=lock['previous']['commit'], mcp_commit=state['mcp_commit'], cases=cases)
    golden = DIRECTORY / 'previous.golden.json'
    if record:
        golden.write_bytes(encode(actual))
    else:
        compare(read_json(golden), actual)
    output = CACHE / 'last-run'
    output.mkdir(exist_ok=True)
    (output / 'upgrade-previous.lua.json').write_bytes(encode(actual))
    print('Эталоны предыдущего официального PoB:', len(cases))


if __name__ == '__main__':
    execute('--record' in sys.argv)
