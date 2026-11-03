"""Smoke must not spend on an active real provider without an explicit flag."""

from io import BytesIO
import json
import sys
from unittest.mock import Mock
import urllib.error

import pytest
import smoke


@pytest.mark.parametrize("simulated,arguments", [(False, []), (True, ["--live"])])
def test_smoke_mode_mismatch_stops_after_readiness(monkeypatch, simulated, arguments):
    monkeypatch.setattr(sys, "argv", ["smoke.py", *arguments])
    response = BytesIO(json.dumps({"simulated": simulated}).encode())
    opener = Mock(return_value=response)
    monkeypatch.setattr(smoke.urllib.request, "urlopen", opener)
    with pytest.raises(SystemExit):
        smoke.main()
    assert opener.call_count == 1


def test_fake_smoke_checks_auth_and_exact_script(monkeypatch):
    monkeypatch.setattr(sys, "argv", ["smoke.py"])
    monkeypatch.setenv("AI_INTERNAL_TOKEN", "test-only-service-token")
    opener = Mock(side_effect=[
        BytesIO(b'{"simulated": true}'),
        BytesIO(b'{"simulated": true, "route": "script", "script_id": "sample"}'),
        urllib.error.HTTPError("http://localhost", 401, "Unauthorized", {}, None),
    ])
    monkeypatch.setattr(smoke.urllib.request, "urlopen", opener)
    smoke.main()
    assert opener.call_count == 3
