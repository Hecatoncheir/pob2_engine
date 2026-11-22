"""Проверки состава и воспроизводимости релизного архива."""

import copy
import hashlib
import json
from pathlib import Path
import tempfile
import unittest
import zipfile

import package_release
import smoke_release


class PackageTests(unittest.TestCase):
    def test_native_binary_does_not_block_linux_cache_directory(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            binary = root / 'pob2-engine'
            binary.write_bytes(b'fixture')
            scenario = smoke_release.Scenario(binary, root)
            cache = Path(scenario.environment['XDG_CACHE_HOME']) / 'pob2-engine' / 'pob'
            cache.mkdir(parents=True)
            self.assertEqual(binary.read_bytes(), b'fixture')
            self.assertTrue(cache.is_dir())

    def test_packages_three_spells_and_published_001_session(self):
        files = package_release.documentation()
        names = ['testdata/reference/fireball.xml',
                 'testdata/reference/0.0.2/cases/frostbolt-minimal.xml',
                 'testdata/reference/0.0.2/cases/spark-minimal.xml',
                 'testdata/compat/0.0.1/manifest.json',
                 'testdata/compat/0.0.1/snapshot.json', 'testdata/compat/0.0.1/sessions.db']
        for name in names:
            self.assertEqual(files[name], (package_release.ROOT / name).read_bytes())

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

    def test_rejects_commit_text_that_disagrees_with_manifest(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            archive = root / 'wrong-commit.zip'
            commit_text = ('b' * 40 + '\n').encode('ascii')
            manifest = {'schema': 1, 'tag': 'v0.0.2-rc.1', 'commit': 'a' * 40,
                        'files': {'COMMIT.txt': hashlib.sha256(commit_text).hexdigest()}}
            with zipfile.ZipFile(archive, 'w') as output:
                output.writestr('PACKAGE.json', json.dumps(manifest))
                output.writestr('COMMIT.txt', commit_text)
            with self.assertRaisesRegex(ValueError, 'COMMIT.txt'):
                smoke_release.unpack(archive, root / 'unpacked', 'v0.0.2-rc.1')

    def test_checks_published_archive_checksum(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            archive, checksums = root / 'release.zip', root / 'SHA256SUMS'
            archive.write_bytes(b'published')
            checksums.write_text(hashlib.sha256(b'published').hexdigest() + '  release.zip\n', 'ascii')
            smoke_release.verify_checksum(archive, checksums)
            archive.write_bytes(b'changed')
            with self.assertRaisesRegex(ValueError, 'SHA256SUMS'):
                smoke_release.verify_checksum(archive, checksums)

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

    def test_upgrade_gate_rejects_changed_inputs_stats_and_unchanged_dataset(self):
        root = package_release.ROOT / 'testdata/compat/0.0.1'
        previous = smoke_release.verify_compatibility_fixture(root)
        updated = copy.deepcopy(previous)
        updated['revision'] = 4
        result = updated['build']['calculation']
        result['data_id'] = smoke_release.DATA_ID
        result['data_sha256'] = 'a' * 64
        smoke_release.check_updated_snapshot(previous, updated)
        for path, value in [(['build', 'xml'], 'изменено'),
                            (['build', 'calculation', 'stats', 'Life'], -1),
                            (['build', 'calculation', 'data_id'], 'pob2-0.0.1-fireball-v1'),
                            (['build', 'calculation', 'data_sha256'], previous['build']['calculation']['data_sha256'])]:
            with self.subTest(path=path):
                invalid = copy.deepcopy(updated)
                target = invalid
                for key in path[:-1]:
                    target = target[key]
                target[path[-1]] = value
                with self.assertRaises(ValueError):
                    smoke_release.check_updated_snapshot(previous, invalid)


if __name__ == '__main__':
    unittest.main()
