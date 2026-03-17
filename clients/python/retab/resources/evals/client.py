from typing import Any

from ..._resource import AsyncAPIResource, SyncAPIResource
from .extract import ExtractProjects, AsyncExtractProjects


class Evals(SyncAPIResource):
    """Evals API — evaluation pipelines for extraction, split, and classification projects."""

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.extract = ExtractProjects(client=client)


class AsyncEvals(AsyncAPIResource):
    """Async Evals API."""

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.extract = AsyncExtractProjects(client=client)
