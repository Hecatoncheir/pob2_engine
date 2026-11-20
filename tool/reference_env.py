"""Изолированные исходники и нативная среда закреплённого Lua-эталона."""

import hashlib
import io
import json
import os
from pathlib import Path
import platform
import struct
import subprocess
import sys
import urllib.request
import zipfile

ROOT = Path(__file__).resolve().parents[1]
CACHE = ROOT / '.reference'
DATA = ROOT / 'testdata/reference'


def sha256(data):
    return hashlib.sha256(data).hexdigest()


def read_json(path):
    return json.loads(path.read_text('utf-8'))


def git(directory, *args):
    return subprocess.check_output(['git', '-c', 'core.autocrlf=false', '-c', 'core.eol=lf', '-C', str(directory), *args],
                                   text=True, encoding='utf-8').strip()


def verify_source(directory, source):
    if git(directory, 'rev-parse', 'HEAD') != source['commit']:
        raise ValueError('Неверная ревизия: ' + str(directory))
    git(directory, 'diff', '--exit-code', '--quiet', '--no-ext-diff', 'HEAD', '--')
    if git(directory, 'ls-files', '--others', '--exclude-standard'):
        raise ValueError('Изменённая рабочая копия эталона: ' + str(directory))
    expected = 'https://github.com/' + source['repository'] + '.git'
    if git(directory, 'remote', 'get-url', 'origin') != expected:
        raise ValueError('Неверный origin: ' + str(directory))
    for entry in source['files']:
        if sha256((directory / entry['path']).read_bytes()) != entry['sha256']:
            raise ValueError('Изменён исходный файл: ' + entry['path'])


def prepare_source(name, source):
    directory = CACHE / name
    if directory.exists():
        verify_source(directory, source)
        return
    directory.mkdir(parents=True)
    git(directory, 'init', '--quiet')
    git(directory, 'config', 'core.autocrlf', 'false')
    git(directory, 'config', 'core.eol', 'lf')
    git(directory, 'remote', 'add', 'origin', 'https://github.com/' + source['repository'] + '.git')
    git(directory, 'fetch', '--quiet', '--depth=1', '--filter=blob:none', 'origin', source['commit'])
    if name == 'pob':
        git(directory, 'sparse-checkout', 'set', 'src', 'runtime')
    git(directory, 'checkout', '--quiet', '--detach', 'FETCH_HEAD')
    verify_source(directory, source)


def target():
    if struct.calcsize('P') != 8 or sys.version_info < (3, 11):
        raise ValueError('Нужен 64-разрядный Python 3.11 или новее')
    machine = platform.machine().lower()
    if sys.platform == 'win32' and machine in {'amd64', 'x86_64'}:
        return 'windows-x86_64'
    return sys.platform + '-' + machine


def native_asset():
    lock = read_json(DATA / 'runtime.lock.json')
    if target() == 'windows-x86_64':
        return None
    if target() not in lock['assets']:
        raise ValueError('Нет закреплённого Lua runtime для ' + target())
    asset = lock['assets'][target()]
    url = 'https://github.com/Hecatoncheir/pob-poe2-mcp/releases/download/'
    return asset, url + lock['release'] + '/' + asset['name']


def unpack_native(data, asset, directory):
    if sha256(data) != asset['sha256']:
        raise ValueError('Неверная SHA-256 нативного архива')
    with zipfile.ZipFile(io.BytesIO(data)) as archive:
        for entry in archive.infolist():
            path = Path(entry.filename)
            if path.is_absolute() or '..' in path.parts or ':' in entry.filename:
                raise ValueError('Недопустимый путь в нативном архиве')
        archive.extractall(directory)
    (directory / 'archive.sha256').write_text(asset['sha256'] + '\n', encoding='ascii')


def native_files():
    if target() == 'windows-x86_64':
        directory = CACHE / 'pob/runtime'
        names = ['lua51.dll', 'lua-utf8.dll']
        return {name: sha256((directory / name).read_bytes()) for name in names}
    asset, _ = native_asset()
    directory = CACHE / target()
    if (directory / 'archive.sha256').read_text('ascii').strip() != asset['sha256']:
        raise ValueError('Неверная версия нативного архива')
    archive_path = directory / asset['name']
    data = archive_path.read_bytes()
    if sha256(data) != asset['sha256']:
        raise ValueError('Изменён нативный архив')
    with zipfile.ZipFile(io.BytesIO(data)) as archive:
        for entry in archive.infolist():
            if not entry.is_dir() and (directory / entry.filename).read_bytes() != archive.read(entry):
                raise ValueError('Изменена нативная библиотека: ' + entry.filename)
    if read_json(directory / 'native/sources.json') != read_json(CACHE / 'mcp/tools/native-lock.json'):
        raise ValueError('Происхождение нативных библиотек не совпало с MCP')
    return {p.name: sha256(p.read_bytes()) for p in sorted((directory / 'native').iterdir()) if p.is_file()}


def prepare():
    lock = read_json(DATA / 'sources.lock.json')
    for name in ['pob', 'mcp']:
        prepare_source(name, lock[name])
    selected = native_asset()
    if selected:
        asset, url = selected
        directory = CACHE / target()
        if not directory.exists():
            with urllib.request.urlopen(url, timeout=60) as response:
                data = response.read(16 * 1024 * 1024 + 1)
            if len(data) > 16 * 1024 * 1024:
                raise ValueError('Нативный архив слишком велик')
            unpack_native(data, asset, directory)
            (directory / asset['name']).write_bytes(data)
    native_files()


def verify():
    lock = read_json(DATA / 'sources.lock.json')
    for name in ['pob', 'mcp']:
        verify_source(CACHE / name, lock[name])
    return {'pob_commit': lock['pob']['commit'], 'mcp_commit': lock['mcp']['commit'],
            'tree_version': lock['tree_version'], 'target': target(),
            'python': platform.python_version(), 'native_files': native_files()}


def environment(home):
    values = dict(os.environ)
    for name in list(values):
        if name.startswith(('POB_', 'LUA_')) or name in {'PYTHONHOME', 'PYTHONPATH'}:
            del values[name]
    values.update(PYTHONPATH=str(CACHE / 'mcp'), PYTHONNOUSERSITE='1',
                  PYTHONDONTWRITEBYTECODE='1', PYTHONUTF8='1',
                  POB_DIR=str(CACHE / 'pob'), POB_MCP_NATIVE_DIR=str(CACHE / target() / 'native'),
                  HOME=str(home), USERPROFILE=str(home), APPDATA=str(home / 'roaming'),
                  LOCALAPPDATA=str(home / 'local'), XDG_DATA_HOME=str(home / 'data'),
                  XDG_CONFIG_HOME=str(home / 'config'), XDG_CACHE_HOME=str(home / 'cache'))
    return values
