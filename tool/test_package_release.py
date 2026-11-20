"""Проверки состава и воспроизводимости релизного архива."""

import hashlib
import json
from pathlib import Path
import tempfile
import unittest
import zipfile

import package_release
import smoke_release


class PackageTests(unittest.TestCase):
    def test_inventory_permissions_and_reproducibility(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            binary = root / 'binary'
            binary.write_bytes(b'fixture')
            args = (binary, 'v0.0.1-rc.1', 'linux', 'amd64', 'a' * 40, root)
            output = package_release.package(*args)
            first = output.read_bytes()
            with zipfile.ZipFile(output) as archive:
                manifest = json.loads(archive.read('PACKAGE.json'))
                self.assertEqual(set(archive.namelist()), set(manifest['files']) | {'PACKAGE.json'})
                for name, digest in manifest['files'].items():
                    self.assertEqual(hashlib.sha256(archive.read(name)).hexdigest(), digest, name)
                for name in ['LICENSE', 'THIRD_PARTY_NOTICES.md', '.licenses/runtime/manifest.json',
                             'internal/gamedata/files/manifest.json', 'internal/calculation/origin.json',
                             'testdata/reference/fireball.xml', 'docs/CLI-BUILDS.md', 'COMMIT.txt']:
                    self.assertIn(name, archive.namelist())
                self.assertEqual(archive.getinfo('pob2-engine').external_attr >> 16, 0o100755)
            self.assertEqual(package_release.package(*args).read_bytes(), first)

    def test_rejects_unchecked_target_and_commit(self):
        with self.assertRaises(ValueError):
            package_release.package(Path('missing'), 'v0.0.1', 'darwin', 'amd64', 'a' * 40, Path('.'))
        with self.assertRaises(ValueError):
            package_release.package(Path('missing'), 'v0.0.1', 'linux', 'amd64', 'short', Path('.'))

    def test_rejects_modified_file_before_execution(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            archive = root / 'modified.zip'
            manifest = {'schema': 1, 'tag': 'v0.0.1', 'files': {'pob2-engine': hashlib.sha256(b'original').hexdigest()}}
            with zipfile.ZipFile(archive, 'w') as output:
                output.writestr('PACKAGE.json', json.dumps(manifest))
                output.writestr('pob2-engine', b'modified')
            with self.assertRaisesRegex(ValueError, 'SHA-256'):
                smoke_release.unpack(archive, root / 'unpacked', 'v0.0.1')

    def test_rejects_path_outside_package(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            archive = root / 'unsafe.zip'
            data = b'escape'
            manifest = {'schema': 1, 'tag': 'v0.0.1', 'files': {'../escape': hashlib.sha256(data).hexdigest()}}
            with zipfile.ZipFile(archive, 'w') as output:
                output.writestr('PACKAGE.json', json.dumps(manifest))
                output.writestr('../escape', data)
            with self.assertRaisesRegex(ValueError, 'Путь вне архива'):
                smoke_release.unpack(archive, root / 'unpacked', 'v0.0.1')
            self.assertFalse((root / 'escape').exists())


if __name__ == '__main__':
    unittest.main()
