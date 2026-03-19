from typing import Any

from ....._resource import AsyncAPIResource, SyncAPIResource
from .....types.projects.model import BuilderDocument, Project
from ..._templates import AsyncTemplatesMixin, SyncTemplatesMixin


class Templates(SyncAPIResource, SyncTemplatesMixin):
    BASE = "/evals/extract"
    PROJECT_MODEL = Project
    BUILDER_DOCUMENT_MODEL = BuilderDocument

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)


class AsyncTemplates(AsyncAPIResource, AsyncTemplatesMixin):
    BASE = "/evals/extract"
    PROJECT_MODEL = Project
    BUILDER_DOCUMENT_MODEL = BuilderDocument

    def __init__(self, client: Any, **kwargs: Any) -> None:
        super().__init__(client=client, **kwargs)
