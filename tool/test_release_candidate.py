import unittest
from unittest.mock import patch

from check_release_candidate import APP, APP_STEPS, APP_WORKFLOW, PLATFORMS
from check_release_candidate import check_app_api, check_app_jobs, check_app_run, stable_tag


CANDIDATE = 'v0.0.17-rc.2'


class ReleaseCandidateTest(unittest.TestCase):
    def test_stable_version_from_candidate(self):
        self.assertEqual(stable_tag(CANDIDATE), 'v0.0.17')
        for invalid in ('v0.0.17', 'v0.0.017-rc.2', 'v0.0.17-rc.0'):
            with self.subTest(invalid=invalid), self.assertRaises(SystemExit):
                stable_tag(invalid)

    def test_run_identifies_full_acceptance_of_exact_candidate(self):
        run = {
            'head_repository': {'full_name': APP},
            'path': APP_WORKFLOW,
            'event': 'workflow_dispatch',
            'display_title': f'Приёмка PoB II с {CANDIDATE}',
        }
        check_app_run(run, CANDIDATE)
        for change in (
            {'display_title': 'Приёмка PoB II с v0.0.17-rc.1'},
            {'display_title': 'Проверки и установщики'},
            {'event': 'pull_request'},
            {'head_repository': {'full_name': 'someone/pob2_ui'}},
        ):
            with self.subTest(change=change), self.assertRaises(SystemExit):
                check_app_run(run | change, CANDIDATE)

    def test_matrix_rejects_missing_or_skipped_platform_checks(self):
        jobs = [{'name': os, 'conclusion': 'success', 'steps': [
            {'name': step, 'conclusion': 'success'} for step in APP_STEPS
        ]} for os in PLATFORMS]
        check_app_jobs(jobs)
        with self.assertRaisesRegex(SystemExit, 'macos-arm64'):
            check_app_jobs(jobs[:-1])
        jobs[1]['steps'][0]['conclusion'] = 'skipped'
        with self.assertRaisesRegex(SystemExit, 'linux-x64'):
            check_app_jobs(jobs)

    @patch('check_release_candidate.git', return_value=b'100644 blob abc\tgen/dart/lib/session.dart\n')
    @patch('check_release_candidate.app_api', return_value={'tree': [
        {'type': 'blob', 'path': 'packages/pob2_engine_api/lib/session.dart', 'sha': 'abc'},
    ]})
    @patch('check_release_candidate.read_url', return_value=f'{CANDIDATE}\n'.encode())
    def test_sdk_and_binary_pin_are_from_candidate(self, read_url, app_api, _):
        check_app_api(CANDIDATE, 'a' * 40)
        app_api.return_value['tree'][0]['sha'] = 'def'
        with self.assertRaisesRegex(SystemExit, 'Dart API'):
            check_app_api(CANDIDATE, 'a' * 40)
        read_url.return_value = b'v0.0.17-rc.1\n'
        with self.assertRaisesRegex(SystemExit, 'закреплён'):
            check_app_api(CANDIDATE, 'a' * 40)
