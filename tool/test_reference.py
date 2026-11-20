"""Проверки целостности эталона и отказов при подмене входа."""

import io
import tempfile
from pathlib import Path
import unittest
import zipfile

from reference_env import DATA, ROOT, read_json, sha256, unpack_native
from reference_lua import compare
from notices import module_files


class ReferenceIntegrity(unittest.TestCase):
    def test_fixture_and_recipe_have_recorded_hashes(self):
        golden = read_json(DATA / 'fireball.lua.json')
        self.assertEqual(golden['input_sha256'], sha256((DATA / 'fireball.xml').read_bytes()))
        self.assertEqual(golden['recipe_sha256'], sha256((DATA / 'fireball.recipe.json').read_bytes()))
        sources = read_json(DATA / 'sources.lock.json')
        self.assertEqual(golden['pob_commit'], sources['pob']['commit'])
        self.assertEqual(golden['mcp_commit'], sources['mcp']['commit'])
        self.assertNotEqual(golden['pob_commit'], sources['mcp_embedded_pob_commit'])

    def test_original_licenses_match_pinned_sources(self):
        sources = read_json(DATA / 'sources.lock.json')
        for project, path in [('pob', 'pob-LICENSE.md'), ('mcp', 'mcp-LICENSE.txt')]:
            entry = next(f for f in sources[project]['files'] if f['path'].startswith('LICENSE'))
            self.assertEqual(sha256((ROOT / '.licenses/reference' / path).read_bytes()), entry['sha256'])

    def test_matrix_inputs_match_recorded_results(self):
        golden = read_json(DATA / 'cases.lua.json')
        self.assertEqual(golden['recipe_sha256'], sha256((DATA / 'cases.recipe.json').read_bytes()))
        recorded = {case['id'] for case in golden['cases']}
        recipes = [case['id'] for case in read_json(DATA / 'cases.recipe.json')['cases']]
        self.assertEqual(len(recipes), len(set(recipes)))
        self.assertEqual(recorded, set(recipes))
        for case in golden['cases']:
            self.assertEqual(case['input_sha256'], sha256((DATA / 'cases' / (case['id'] + '.xml')).read_bytes()))

    def test_spell_matrix_inputs_and_provenance(self):
        directory = DATA / '0.0.2'
        golden = read_json(directory / 'spells.lua.json')
        recipe = read_json(directory / 'spells.recipe.json')
        self.assertEqual(golden['recipe_sha256'], sha256((directory / 'spells.recipe.json').read_bytes()))
        sources = read_json(DATA / 'sources.lock.json')
        self.assertEqual(golden['pob_commit'], sources['pob']['commit'])
        self.assertEqual(golden['mcp_commit'], sources['mcp']['commit'])
        recipes = {case['id']: case for case in recipe['cases']}
        self.assertEqual(len(recipe['cases']), len(recipes))
        self.assertEqual({case['id'] for case in golden['cases']}, set(recipes))
        self.assertEqual(len(golden['cases']), 84)
        for case in golden['cases']:
            self.assertEqual(case['input_sha256'], sha256((directory / 'cases' / (case['id'] + '.xml')).read_bytes()))
            self.assertEqual(case['skill'], recipes[case['id']]['skill'])
            self.assertEqual(case['element'], recipes[case['id']]['element'])

    def test_warnings_and_field_loss_are_not_numeric_tolerances(self):
        for actual in [{'warnings': []}, {'warnings': ['different']}, {}]:
            with self.subTest(actual=actual), self.assertRaises(ValueError):
                compare({'warnings': ['unsupported modifier']}, actual)

    def test_booleans_and_significant_numeric_differences_fail(self):
        for expected, actual in [(True, 1), (0, 0.01), (1.0, float('nan'))]:
            with self.subTest(actual=actual), self.assertRaises(ValueError):
                compare(expected, actual)
        compare({'dps': 100.0}, {'dps': 100.000000001})

    def test_native_archive_hash_is_checked_before_writing(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary) / 'native'
            with self.assertRaisesRegex(ValueError, 'SHA-256'):
                unpack_native(b'bad data', {'sha256': '0' * 64}, directory)
            self.assertFalse(directory.exists())

    def test_native_archive_cannot_escape_destination(self):
        data = io.BytesIO()
        with zipfile.ZipFile(data, 'w') as archive:
            archive.writestr('../escaped', 'bad data')
        content = data.getvalue()
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary) / 'native'
            with self.assertRaisesRegex(ValueError, 'путь'):
                unpack_native(content, {'sha256': sha256(content)}, directory)
            self.assertFalse((Path(temporary) / 'escaped').exists())

    def test_missing_module_license_blocks_distribution(self):
        with tempfile.TemporaryDirectory() as temporary:
            with self.assertRaisesRegex(ValueError, 'лицензии'):
                module_files({'Dir': temporary, 'Path': 'example.org/no-license'})


if __name__ == '__main__':
    unittest.main()
