"""Typed exception hierarchy for the Retab SDK.

All API exceptions inherit from both :class:`APIError` and :class:`RuntimeError`,
so existing ``except RuntimeError`` blocks continue to catch them.
"""

from typing import Any


class RetabError(Exception):
    """Base class for all Retab SDK errors."""

    def __init__(self, message: str) -> None:
        self.message = message
        super().__init__(message)


class APIError(RetabError, RuntimeError):
    """A non-success HTTP response from the Retab API.

    Also inherits from :class:`RuntimeError` for backward compatibility
    with code that catches ``RuntimeError``.
    """

    def __init__(
        self,
        message: str,
        status_code: int,
        code: str | None = None,
        details: dict[str, Any] | None = None,
        body: str = "",
        request_id: str | None = None,
    ) -> None:
        self.status_code = status_code
        self.code = code
        self.details = details
        self.body = body
        self.request_id = request_id
        super().__init__(message)

    def __repr__(self) -> str:
        return (
            f"{self.__class__.__name__}(message={self.message!r}, "
            f"status_code={self.status_code}, code={self.code!r}, "
            f"request_id={self.request_id!r})"
        )


class AuthenticationError(APIError):
    """HTTP 401 — invalid or missing API key."""
    pass


class PermissionDeniedError(APIError):
    """HTTP 403 — valid API key but insufficient permissions."""
    pass


class NotFoundError(APIError):
    """HTTP 404 — the requested resource does not exist."""
    pass


class ValidationError(APIError):
    """HTTP 422 — request body failed server-side validation."""
    pass


class RateLimitError(APIError):
    """HTTP 429 — too many requests."""
    pass


class InternalServerError(APIError):
    """HTTP 5xx — server-side error."""
    pass


class APIConnectionError(RetabError):
    """Network-level error (DNS failure, refused connection, etc.)."""
    pass


class APITimeoutError(APIConnectionError):
    """The request timed out."""
    pass
