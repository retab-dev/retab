"""Shared sample payloads for tests.

Tiny, deterministic fixtures used to build requests without touching the network
(an inline data-URL document, etc.). Import these instead of re-declaring the
same ``_sample_document`` helper in every unit suite.
"""

from __future__ import annotations

from typing import Any

from retab.types.mime import MIMEData

# A minimal inline document as a base64 data-URL ("invoice"). It is never
# uploaded or processed — it just gives request-shape tests a valid document.
SAMPLE_DOCUMENT_URL = "data:text/plain;base64,aW52b2ljZQ=="


def sample_document() -> MIMEData:
    """A tiny inline ``MIMEData`` document for request-shape / serialization tests."""
    return MIMEData(filename="invoice.txt", url=SAMPLE_DOCUMENT_URL)


def list_envelope(*items: dict[str, Any], before: str | None = None, after: str | None = None) -> dict[str, Any]:
    """A paginated list wire envelope: ``{"data": [...], "list_metadata": {...}}``.

    ``after`` defaults to ``None`` (terminal page). Set it to a cursor string to
    simulate a page that has more results.
    """
    return {"data": list(items), "list_metadata": {"before": before, "after": after}}
