"""Упаковать самостоятельный движок с документацией и исходными уведомлениями."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import zipfile

from notices import archive as license_archive

ROOT = Path(__file__).resolve().parents[1]
TARGETS = {('windows', 'amd64'), ('linux', 'amd64'), ('darwin', 'arm64')}


def documentation():
    selected = {p.resolve() for p in ROOT.glob('*.md')}
    selected.update(p.resolve() for p in (ROOT / 'docs').glob('*.md'))
    selected.update(p.resolve() for p in (ROOT / '.licenses').rglob('*') if p.is_file())
    selected.update(p.resolve() for p in (ROOT / 'internal/gamedata/files').glob('*.json'))
    selected.add((ROOT / 'internal/calculation/origin.json').resolve())
    selected.add((ROOT / 'testdata/reference/fireball.xml').resolve())
    pending = list(selected)
    while pending:
        path = pending.pop()
        if path.suffix != '.md':
            continue
        for target in re.findall(r'\]\(([^)]+)\)', path.read_text('utf-8')):
            if '://' in target or target.startswith('#'):
                continue
            linked = (path.parent / target.split('#', 1)[0]).resolve()
            if not linked.is_relative_to(ROOT) or not linked.is_file():
                raise ValueError('Неверная ссылка документации: ' + target)
            if linked not in selected:
                selected.add(linked)
                pending.append(linked)
    return {p.relative_to(ROOT).as_posix(): p.read_bytes() for p in sorted(selected)}


def package(binary, tag, goos, arch, commit, destination):
    if (goos, arch) not in TARGETS or not re.fullmatch(r'v\d+\.\d+\.\d+(?:-[\w.-]+)?', tag):
        raise ValueError('Неверная платформа или тег')
    if not re.fullmatch(r'[0-9a-f]{40}', commit):
        raise ValueError('Нужен полный commit SHA')
    destination.mkdir(parents=True, exist_ok=True)
    # Проверка оригинальных байтов лицензий перед включением в каждый архив.
    license_archive(destination / f'pob2_engine_{tag}_licenses.zip')
    files = documentation()
    files['LICENSE'] = (ROOT / 'LICENSE').read_bytes()
    executable = 'pob2-engine.exe' if goos == 'windows' else 'pob2-engine'
    files[executable] = binary.read_bytes()
    if not files[executable]:
        raise ValueError('Пустой исполняемый файл')
    files['COMMIT.txt'] = (commit + '\n').encode('ascii')
    manifest = {'schema': 1, 'tag': tag, 'commit': commit, 'goos': goos, 'goarch': arch,
                'executable': executable, 'files': {name: hashlib.sha256(data).hexdigest()
                                                   for name, data in sorted(files.items())}}
    files['PACKAGE.json'] = (json.dumps(manifest, ensure_ascii=False, indent=2) + '\n').encode('utf-8')
    output = destination / f'pob2_engine_{tag}_{goos}_{arch}.zip'
    with zipfile.ZipFile(output, 'w', compression=zipfile.ZIP_DEFLATED) as archive:
        for name, data in sorted(files.items()):
            info = zipfile.ZipInfo(name, (1980, 1, 1, 0, 0, 0))
            info.compress_type = zipfile.ZIP_DEFLATED
            info.external_attr = (0o100755 if name == executable else 0o100644) << 16
            archive.writestr(info, data)
    return output


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--tag', required=True)
    parser.add_argument('--os', required=True)
    parser.add_argument('--arch', required=True)
    parser.add_argument('--commit', required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    print(package(args.binary, args.tag, args.os, args.arch, args.commit, args.output))


if __name__ == '__main__':
    main()
