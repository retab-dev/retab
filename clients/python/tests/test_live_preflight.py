from email.message import Message
from urllib.error import HTTPError, URLError

import pytest

import conftest


class FakeResponse:
    status = 200

    def __enter__(self) -> "FakeResponse":
        return self

    def __exit__(self, exc_type: object, exc: object, traceback: object) -> None:
        return None

    def read(self, _limit: int = -1) -> bytes:
        return b"{}"


@pytest.mark.unit
def test_live_api_preflight_accepts_authenticated_response(monkeypatch: pytest.MonkeyPatch) -> None:
    seen_urls: list[str] = []

    def fake_urlopen(request: object, timeout: int) -> FakeResponse:
        seen_urls.append(request.full_url)  # type: ignore[attr-defined]
        assert timeout == 4
        return FakeResponse()

    monkeypatch.setattr(conftest, "urlopen", fake_urlopen)

    status, reason = conftest._live_api_preflight("sk_valid", "http://localhost:4000/v1")

    assert status == "ok"
    assert reason == ""
    assert seen_urls == ["http://localhost:4000/v1/files?limit=1"]


@pytest.mark.unit
def test_live_api_preflight_reports_invalid_key(monkeypatch: pytest.MonkeyPatch) -> None:
    def fake_urlopen(_request: object, timeout: int) -> FakeResponse:
        assert timeout == 4
        raise HTTPError("http://localhost:4000/v1/files?limit=1", 401, "Unauthorized", Message(), None)

    monkeypatch.setattr(conftest, "urlopen", fake_urlopen)

    status, reason = conftest._live_api_preflight("sk_bad", "http://localhost:4000")

    assert status == "invalid_key"
    assert "RETAB_API_KEY" in reason


@pytest.mark.unit
def test_live_api_preflight_skips_unreachable_server(monkeypatch: pytest.MonkeyPatch) -> None:
    def fake_urlopen(_request: object, timeout: int) -> FakeResponse:
        assert timeout == 4
        raise URLError("connection refused")

    monkeypatch.setattr(conftest, "urlopen", fake_urlopen)

    status, reason = conftest._live_api_preflight("sk_valid", "http://localhost:4000")

    assert status == "unreachable"
    assert "unreachable" in reason
