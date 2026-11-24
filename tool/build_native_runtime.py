"""Собрать библиотеки Lua из закреплённых исходников для бинарника движка."""

import json
import os
from pathlib import Path
import platform
import subprocess
import sys
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parents[1]
LOCK = ROOT / 'tool/native-runtime.lock.json'


def run(*arguments):
    subprocess.run(arguments, check=True)


def checkout(root, name, source):
    target = root / name
    run('git', 'init', '--quiet', str(target))
    run('git', '-C', str(target), 'fetch', '--depth=1', source['repository'], source['revision'])
    run('git', '-C', str(target), 'checkout', '--detach', 'FETCH_HEAD')
    return target


def build(root, sources, darwin):
    jit = checkout(root, 'luajit', sources['luajit'])
    utf8 = checkout(root, 'luautf8', sources['luautf8'])
    library = 'libluajit-5.1.dylib' if darwin else 'libluajit-5.1.so.2'
    module = root / 'lua-utf8.so'
    run('make', '-C', str(jit), '-j2')
    options = ['-bundle', '-undefined', 'dynamic_lookup'] if darwin else ['-shared', '-fPIC']
    run('cc', *options, '-O2', '-I' + str(jit / 'src'), str(utf8 / 'lutf8lib.c'), '-o', str(module))
    return {library: jit / 'src/libluajit.so', 'lua-utf8.so': module,
            'LICENSE-LuaJIT.txt': jit / 'COPYRIGHT', 'LICENSE-luautf8.txt': utf8 / 'LICENSE',
            'sources.json': LOCK}


def main():
    darwin = sys.platform == 'darwin'
    target = {('linux', 'x86_64'): 'linux-amd64', ('darwin', 'arm64'): 'darwin-arm64'}
    platform_name = target.get((sys.platform, platform.machine()))
    if platform_name is None:
        raise SystemExit('Для выпуска нужны нативные Linux x64 или macOS arm64')
    if darwin:
        os.environ.setdefault('MACOSX_DEPLOYMENT_TARGET', '11.0')
    output = ROOT / 'internal/pob/runtime' / (platform_name + '.zip')
    sources = json.loads(LOCK.read_text(encoding='utf-8'))
    with tempfile.TemporaryDirectory(prefix='pob2-lua-native-') as temporary:
        files = build(Path(temporary), sources, darwin)
        with zipfile.ZipFile(output, 'w', zipfile.ZIP_DEFLATED) as archive:
            for name, path in files.items():
                archive.write(path, 'native/' + name)
    print(output)


if __name__ == '__main__':
    main()
