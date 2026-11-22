"""Сопоставить CLI и gRPC с памятью и диском на двух выпущенных бинарниках."""

import hashlib
import json
from pathlib import Path
import platform
import subprocess
import sys
import tempfile

from reference_env import CACHE, ROOT, verify
from smoke_multigroup import Matrix
from smoke_release import Scenario, require, verify_checksum


def check_source(result):
    require(result['data_id'] == 'pob-lua-local', 'Не выбран закреплённый локальный PoB')


def grpc_measure(probe, ready):
    result = subprocess.run([str(probe)], input=ready, capture_output=True, check=True, timeout=80)
    return json.loads(result.stdout)


def memory_measure(scenario, probe):
    with subprocess.Popen([str(scenario.binary), 'serve', '--ephemeral'], cwd=scenario.root,
                          env=scenario.environment, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL) as server:
        try:
            return grpc_measure(probe, server.stdout.readline())
        finally:
            if server.poll() is None:
                server.terminate()
                server.wait(timeout=10)


def durable_measure(scenario, probe):
    scenario.start('direct')
    record = scenario.root / 'run/pob2-engine/direct/connection.json'
    result = grpc_measure(probe, record.read_bytes())
    scenario.stop()
    return result


def measure(binary, probe):
    with tempfile.TemporaryDirectory(prefix='measure-upgrade-', dir=CACHE) as directory:
        scenario = Scenario(binary, Path(directory))
        scenario.environment.pop('POB2_ENGINE_CALCULATOR')
        scenario.environment['POB2_ENGINE_POB_DIR'] = str(CACHE / 'pob')
        try:
            version = scenario.run('--version')['version']
            scenario.start('cli')
            cli = Matrix(scenario, ROOT, check_source).measure()
            scenario.stop()
            disk = durable_measure(scenario, probe)
            memory = memory_measure(scenario, probe)
        finally:
            scenario.close()
    digest = hashlib.sha256(binary.read_bytes()).hexdigest()
    return dict(version=version, binary_sha256=digest, cli=cli, grpc_disk=disk, grpc_memory=memory)


def main():
    old, new, probe = (Path(path).resolve() for path in sys.argv[1:4])
    for binary in (old, new):
        verify_checksum(binary, binary.parent / 'SHA256SUMS')
    state = verify()
    passes = [measure(binary, probe) for binary in (old, new, new, old)]
    result = dict(schema=1, system=platform.system(), architecture=platform.machine(), os_version=platform.version(),
                  pob_commit=state['pob_commit'], sample_count=5, passes=passes)
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
