from typing import Any
from ...._resource import SyncAPIResource, AsyncAPIResource
from .documentai import DocumentAIs, AsyncDocumentAIs

class Templates(SyncAPIResource): 
    """Templates API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.documentai = DocumentAIs(client=client)


class AsyncTemplates(AsyncAPIResource):
    """Templates API wrapper for asynchronous usage."""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.documentai = AsyncDocumentAIs(client=client)
