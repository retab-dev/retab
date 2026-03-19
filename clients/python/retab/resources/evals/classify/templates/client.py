from typing import Any

from ....._resource import AsyncAPIResource, SyncAPIResource
from .....types.evals.classify import ClassifyBuilderDocument, ClassifyProject
from ..._templates import AsyncTemplatesMixin, SyncTemplatesMixin


class Templates(SyncAPIResource, SyncTemplatesMixin):
    BASE = "/evals/classify"
    PROJECT_MODEL = ClassifyProject
    BUILDER_DOCUMENT_MODEL = ClassifyBuilderDocument

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)


class AsyncTemplates(AsyncAPIResource, AsyncTemplatesMixin):
    BASE = "/evals/classify"
    PROJECT_MODEL = ClassifyProject
    BUILDER_DOCUMENT_MODEL = ClassifyBuilderDocument

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
