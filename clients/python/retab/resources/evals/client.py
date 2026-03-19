from typing import Any

from ..._resource import AsyncAPIResource, SyncAPIResource
from .classify import AsyncClassify, Classify
from .extract import AsyncExtract, Extract
from .split import AsyncSplit, Split


class Evals(SyncAPIResource):
    """Evals API — evaluation pipelines for extraction, split, and classification projects."""

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.extract = Extract(client=client)
        self.split = Split(client=client)
        self.classify = Classify(client=client)


class AsyncEvals(AsyncAPIResource):
    """Async Evals API."""

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
        self.extract = AsyncExtract(client=client)
        self.split = AsyncSplit(client=client)
        self.classify = AsyncClassify(client=client)
