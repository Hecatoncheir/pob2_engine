"""Собрать оригинальные уведомления зависимостей CLI для трёх целевых платформ."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile

ROOT = Path(__file__).resolve().parents[1]
OUTPUT = ROOT / '.licenses/runtime'
TARGETS = [('windows', 'amd64'), ('linux', 'amd64'), ('darwin', 'arm64')]
LEGAL_NAMES = {'LICENSE', 'LICENSE.md', 'LICENSE.txt', 'COPYING', 'COPYRIGHT',
               'NOTICE', 'NOTICE.txt', 'AUTHORS', 'PATENTS'}


def run(*args, environment=None):
    return subprocess.check_output(args, cwd=ROOT, env=environment, text=True, encoding='utf-8')


def objects(text):
    decoder = json.JSONDecoder()
    position = 0
    while position < len(text):
        if text[position].isspace():
            position += 1
            continue
        value, position = decoder.raw_decode(text, position)
        yield value


def legal_files(root):
    return sorted(p for p in root.iterdir() if p.is_file() and
                  (p.name in LEGAL_NAMES or p.name.startswith('LICENSE-')))


def module_files(module):
    directory = Path(module['Dir'])
    files = legal_files(directory)
    if not any(p.name.startswith(('LICENSE', 'COPYING', 'COPYRIGHT')) for p in files):
        raise ValueError('Нет исходного текста лицензии: ' + module['Path'])
    return files


def collect():
    run('go', 'mod', 'verify')
    modules = {}
    for goos, arch in TARGETS:
        environment = dict(os.environ, GOOS=goos, GOARCH=arch, CGO_ENABLED='0', GOWORK='off')
        for package in objects(run('go', 'list', '-deps', '-json', './cmd/pob2-engine', environment=environment)):
            module = package.get('Module')
            if not module or module.get('Main'):
                continue
            if module.get('Replace'):
                raise ValueError('Нужно явно проверить лицензию заменённого модуля: ' + module['Path'])
            modules[module['Path']] = module
    return modules


def bundle(modules, goroot, version):
    files = {}
    inventory = []
    for name, module in sorted(modules.items()):
        selected = []
        for original in module_files(module):
            relative = '/'.join(['modules', name, module['Version'], original.name])
            files[relative] = original.read_bytes()
            selected.append(relative)
        inventory.append({'module': name, 'version': module['Version'], 'files': selected})
    # Уведомления стандартной библиотеки и поставляемых с Go vendor-пакетов.
    for original in legal_files(goroot):
        files['stdlib/' + original.name] = original.read_bytes()
    vendor = goroot / 'src/vendor'
    for original in sorted(vendor.rglob('*')):
        if original.is_file() and original.name in LEGAL_NAMES:
            files['stdlib/vendor/' + original.relative_to(vendor).as_posix()] = original.read_bytes()
    if 'stdlib/LICENSE' not in files:
        raise ValueError('Нет лицензии стандартной библиотеки Go')
    manifest = {'schema': 1, 'go_version': version, 'targets': [a + '/' + b for a, b in TARGETS],
                'modules': inventory, 'files': {name: hashlib.sha256(data).hexdigest()
                                                for name, data in sorted(files.items())}}
    files['manifest.json'] = (json.dumps(manifest, ensure_ascii=False, indent=2) + '\n').encode('utf-8')
    return files


def synchronize(files, write):
    for name, data in sorted(files.items()):
        target = OUTPUT / name
        if write:
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(data)
        elif not target.is_file() or target.read_bytes() != data:
            raise ValueError('Уведомления устарели: ' + str(target.relative_to(ROOT)))
    extra = {p.relative_to(OUTPUT).as_posix() for p in OUTPUT.rglob('*') if p.is_file()} - set(files)
    if extra:
        raise ValueError('Проверьте устаревшие уведомления вручную: ' + ', '.join(sorted(extra)))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--write', action='store_true', help='обновить копии оригинальных уведомлений')
    parser.add_argument('--archive', type=Path, help='собрать комплект лицензий без запуска Go')
    args = parser.parse_args()
    if args.archive:
        archive(args.archive)
        return
    version = run('go', 'env', 'GOVERSION').strip()
    expected = next(line.split()[1] for line in (ROOT / 'go.mod').read_text('utf-8').splitlines() if line.startswith('go '))
    if version != 'go' + expected:
        raise ValueError('Нужен Go из go.mod: ' + expected)
    modules = collect()
    files = bundle(modules, Path(run('go', 'env', 'GOROOT').strip()), version)
    synchronize(files, args.write)
    print('Модулей CLI:', len(modules), '; файлов уведомлений:', len(files) - 1)


def archive(destination):
    manifest = json.loads((OUTPUT / 'manifest.json').read_text('utf-8'))
    for name, digest in manifest['files'].items():
        if hashlib.sha256((OUTPUT / name).read_bytes()).hexdigest() != digest:
            raise ValueError('Повреждено уведомление: ' + name)
    selected = [ROOT / 'LICENSE', ROOT / 'THIRD_PARTY_NOTICES.md']
    selected += sorted(p for p in (ROOT / '.licenses').rglob('*') if p.is_file())
    destination.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(destination, 'w', compression=zipfile.ZIP_DEFLATED) as output:
        for path in selected:
            info = zipfile.ZipInfo(path.relative_to(ROOT).as_posix(), (1980, 1, 1, 0, 0, 0))
            info.compress_type = zipfile.ZIP_DEFLATED
            info.external_attr = 0o100644 << 16
            output.writestr(info, path.read_bytes())
    print('Комплект лицензий:', destination)


if __name__ == '__main__':
    main()
