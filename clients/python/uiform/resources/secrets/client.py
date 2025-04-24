from typing import Any

from ..._resource import AsyncAPIResource, SyncAPIResource
from .external_api_keys import AsyncExternalAPIKeys, ExternalAPIKeys


class Secrets(SyncAPIResource):
    """Automations API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.external_api_keys = ExternalAPIKeys(client=client)


class AsyncSecrets(AsyncAPIResource):
    """Automations API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.external_api_keys = AsyncExternalAPIKeys(client=client)
