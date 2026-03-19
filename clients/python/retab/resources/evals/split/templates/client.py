from typing import Any

from ....._resource import AsyncAPIResource, SyncAPIResource
from .....types.evals.split import SplitBuilderDocument, SplitProject
from ..._templates import AsyncTemplatesMixin, SyncTemplatesMixin


class Templates(SyncAPIResource, SyncTemplatesMixin):
    BASE = "/evals/split"
    PROJECT_MODEL = SplitProject
    BUILDER_DOCUMENT_MODEL = SplitBuilderDocument

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)


class AsyncTemplates(AsyncAPIResource, AsyncTemplatesMixin):
    BASE = "/evals/split"
    PROJECT_MODEL = SplitProject
    BUILDER_DOCUMENT_MODEL = SplitBuilderDocument

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
