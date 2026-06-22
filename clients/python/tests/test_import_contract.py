import json
import subprocess
import sys

import pytest

pytestmark = pytest.mark.unit


def _run_import_probe(source: str) -> dict[str, object]:
    result = subprocess.run(
        [sys.executable, "-c", source],
        capture_output=True,
        text=True,
        timeout=30,
    )
    assert result.returncode == 0, result.stdout + result.stderr
    marker = next(
        (line for line in result.stdout.splitlines() if line.startswith("___PROBE___")),
        None,
    )
    assert marker is not None, result.stdout + result.stderr
    return json.loads(marker[len("___PROBE___") :])


def test_type_submodule_import_does_not_import_client_resources() -> None:
    payload = _run_import_probe(
        """
import json
import sys

from retab.types.mime import MIMEData

print("___PROBE___" + json.dumps({
    "mime_data_name": MIMEData.__name__,
    "retab_client_loaded": "retab.client" in sys.modules,
    "retab_resources_loaded": any(
        name == "retab.resources" or name.startswith("retab.resources.")
        for name in sys.modules
    ),
}))
"""
    )

    assert payload["mime_data_name"] == "MIMEData"
    assert payload["retab_client_loaded"] is False
    assert payload["retab_resources_loaded"] is False


def test_top_level_client_exports_still_work() -> None:
    payload = _run_import_probe(
        """
import json
import sys

from retab import AsyncRetab, MIMEData, Retab

print("___PROBE___" + json.dumps({
    "retab_name": Retab.__name__,
    "async_retab_name": AsyncRetab.__name__,
    "mime_data_name": MIMEData.__name__,
    "retab_client_loaded": "retab.client" in sys.modules,
}))
"""
    )

    assert payload["retab_name"] == "Retab"
    assert payload["async_retab_name"] == "AsyncRetab"
    assert payload["mime_data_name"] == "MIMEData"
    assert payload["retab_client_loaded"] is True
