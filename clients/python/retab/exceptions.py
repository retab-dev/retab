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
        method: str | None = None,
        url: str | None = None,
    ) -> None:
        self.status_code = status_code
        self.code = code
        self.details = details
        self.body = body
        self.request_id = request_id
        self.method = method
        self.url = url
        self.retries: int = 0
        super().__init__(message)

    def __str__(self) -> str:
        lines = [f"{self.status_code} — {self.message}"]
        if self.method and self.url:
            lines.append(f"  URL:        {self.method} {self.url}")
        if self.request_id:
            lines.append(f"  Request-ID: {self.request_id}")
        if self.code:
            lines.append(f"  Code:       {self.code}")
        if self.details:
            lines.append(f"  Details:    {self.details}")
        if self.body:
            truncated = self.body[:500]
            if len(self.body) > 500:
                truncated += "..."
            lines.append(f"  Body:       {truncated}")
        if self.retries > 0:
            lines.append(f"  Retries:    {self.retries}")
        return "\n".join(lines)

    def __repr__(self) -> str:
        parts = [
            f"message={self.message!r}",
            f"status_code={self.status_code}",
            f"code={self.code!r}",
            f"request_id={self.request_id!r}",
        ]
        if self.method and self.url:
            parts.append(f"url={self.method!r} {self.url!r}")
        if self.details:
            parts.append(f"details={self.details!r}")
        if self.body:
            parts.append(f"body={self.body!r}")
        if self.retries > 0:
            parts.append(f"retries={self.retries}")
        return f"{self.__class__.__name__}({', '.join(parts)})"


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
