"""Offline mock harness for unit tests.

Builds a real ``Retab`` / ``AsyncRetab`` client whose transport
(``_prepared_request``) is replaced by a recorder that captures the
``PreparedRequest`` the SDK would send and returns a canned response — no
network, no credentials. Use it to assert request *shape* (route, params, body)
and response deserialization without a server.

This replaces the per-test ``captured = {}`` + ``def fake_prepared_request`` +
``monkeypatch.setattr(client, "_prepared_request", ...)`` boilerplate that the
unit suites used to repeat dozens of times.

Usage::

    from mocks import mock_retab

    # canned response served on every call
    client, rec = mock_retab({"id": "extr_123", ...})
    with client:
        result = client.extractions.create(...)
    assert rec.request.url == "/v1/extractions"          # last request
    assert rec.request.data["document"] == {...}

    # route-aware responder (multiple calls)
    client, rec = mock_retab(
        lambda req: SESSION if req.url == "/v1/files/upload" else MIME
    )
    with client:
        client.files.create_upload(...); client.files.complete_upload(...)
    assert [r.url for r in rec.requests] == [...]
"""

from __future__ import annotations

from typing import Any, Callable, Union

from retab import AsyncRetab, Retab
from retab.types.standards import PreparedRequest

# A responder is either a fixed value returned on every call, or a callable that
# receives the PreparedRequest and returns the response (for route-aware logic).
Responder = Union[Any, Callable[[PreparedRequest], Any]]


class RequestRecorder:
    """Records every ``PreparedRequest`` and serves the configured response."""

    def __init__(self, responder: Responder) -> None:
        self.requests: list[PreparedRequest] = []
        self._responder = responder

    def __call__(self, request: PreparedRequest) -> Any:
        self.requests.append(request)
        responder = self._responder
        return responder(request) if callable(responder) else responder

    @property
    def request(self) -> PreparedRequest:
        """The most recent (or only) captured request."""
        if not self.requests:
            raise AssertionError("no request was captured by the mock client")
        return self.requests[-1]


def mock_retab(
    responder: Responder = None,
    *,
    base_url: str = "http://example.com",
    api_key: str = "test",
) -> tuple[Retab, RequestRecorder]:
    """Return ``(client, recorder)`` for a sync client with a stubbed transport.

    ``responder`` defaults to ``{}`` (empty response). The caller owns the client
    lifecycle — use ``with client:`` or ``client.close()``.
    """
    client = Retab(api_key=api_key, base_url=base_url)
    recorder = RequestRecorder({} if responder is None else responder)
    client._prepared_request = recorder  # type: ignore[method-assign]
    return client, recorder


def mock_async_retab(
    responder: Responder = None,
    *,
    base_url: str = "http://example.com",
    api_key: str = "test",
) -> tuple[AsyncRetab, RequestRecorder]:
    """Async counterpart of :func:`mock_retab` (awaitable transport)."""
    client = AsyncRetab(api_key=api_key, base_url=base_url)
    recorder = RequestRecorder({} if responder is None else responder)

    async def _async_transport(request: PreparedRequest) -> Any:
        return recorder(request)

    client._prepared_request = _async_transport  # type: ignore[method-assign]
    return client, recorder
