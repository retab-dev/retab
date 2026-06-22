from importlib import import_module
from typing import TYPE_CHECKING, Any

__all__ = [
    "Retab",
    "AsyncRetab",
    "utils",
    "types",
    # Exceptions
    "RetabError",
    "APIError",
    "AuthenticationError",
    "PermissionDeniedError",
    "NotFoundError",
    "ConflictError",
    "ValidationError",
    "RateLimitError",
    "InternalServerError",
    "APIConnectionError",
    "APITimeoutError",
    # Response types
    "Classification",
    "Partition",
    "Split",
    "Extraction",
    "ExtractionRequest",
    "SourcesResponse",
    # Core types
    "MIMEData",
    "PaginatedList",
    "AsyncPaginatedList",
]

_LAZY_EXPORTS: dict[str, tuple[str, str | None]] = {
    "Retab": (".client", "Retab"),
    "AsyncRetab": (".client", "AsyncRetab"),
    "utils": (".utils", None),
    "types": (".types", None),
    "RetabError": (".exceptions", "RetabError"),
    "APIError": (".exceptions", "APIError"),
    "AuthenticationError": (".exceptions", "AuthenticationError"),
    "PermissionDeniedError": (".exceptions", "PermissionDeniedError"),
    "NotFoundError": (".exceptions", "NotFoundError"),
    "ConflictError": (".exceptions", "ConflictError"),
    "ValidationError": (".exceptions", "ValidationError"),
    "RateLimitError": (".exceptions", "RateLimitError"),
    "InternalServerError": (".exceptions", "InternalServerError"),
    "APIConnectionError": (".exceptions", "APIConnectionError"),
    "APITimeoutError": (".exceptions", "APITimeoutError"),
    "Classification": (".types.classifications", "Classification"),
    "Partition": (".types.partitions", "Partition"),
    "Split": (".types.splits", "Split"),
    "Extraction": (".types.extractions", "Extraction"),
    "ExtractionRequest": (".types.extractions", "ExtractionRequest"),
    "SourcesResponse": (".types.extractions", "SourcesResponse"),
    "MIMEData": (".types.mime", "MIMEData"),
    "PaginatedList": (".types.pagination", "PaginatedList"),
    "AsyncPaginatedList": (".types.pagination", "AsyncPaginatedList"),
}


def __getattr__(name: str) -> Any:
    try:
        module_name, attribute_name = _LAZY_EXPORTS[name]
    except KeyError as exc:
        raise AttributeError(f"module {__name__!r} has no attribute {name!r}") from exc

    module = import_module(module_name, __name__)
    value = module if attribute_name is None else getattr(module, attribute_name)
    globals()[name] = value
    return value


def __dir__() -> list[str]:
    return sorted([*globals(), *__all__])


if TYPE_CHECKING:
    from . import types, utils
    from .client import AsyncRetab, Retab
    from .exceptions import (
        APIConnectionError,
        APIError,
        APITimeoutError,
        AuthenticationError,
        ConflictError,
        InternalServerError,
        NotFoundError,
        PermissionDeniedError,
        RateLimitError,
        RetabError,
        ValidationError,
    )
    from .types.classifications import Classification
    from .types.extractions import Extraction, ExtractionRequest, SourcesResponse
    from .types.mime import MIMEData
    from .types.pagination import AsyncPaginatedList, PaginatedList
    from .types.partitions import Partition
    from .types.splits import Split
