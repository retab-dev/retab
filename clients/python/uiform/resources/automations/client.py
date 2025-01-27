from typing import Any

from ..._resource import SyncAPIResource, AsyncAPIResource

from .emails import Emails
from .extraction_link import ExtractionLink



class Automations(SyncAPIResource): 
    """Automations API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.emails = Emails(client=client)
        self.extraction_link = ExtractionLink(client=client)


