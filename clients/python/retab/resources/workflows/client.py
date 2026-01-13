from typing import Any

from ..._resource import AsyncAPIResource, SyncAPIResource
from .runs import WorkflowRuns, AsyncWorkflowRuns


class Workflows(SyncAPIResource):
    """Workflows API wrapper for synchronous operations.

    Sub-clients:
        runs: Workflow run operations (create, get)
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.runs = WorkflowRuns(client=client)


class AsyncWorkflows(AsyncAPIResource):
    """Workflows API wrapper for asynchronous operations.

    Sub-clients:
        runs: Workflow run operations (create, get)
    """

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.runs = AsyncWorkflowRuns(client=client)
