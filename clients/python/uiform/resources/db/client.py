from typing import Any

from ..._resource import SyncAPIResource, AsyncAPIResource

from .files import Files, AsyncFiles
from .dataset_memberships import DatasetMemberships, AsyncDatasetMemberships
from .datasets import Datasets, AsyncDatasets
from .annotations import Annotations, AsyncAnnotations



class DB(SyncAPIResource): 
    """DB API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.files = Files(client=client)
        self.dataset_memberships = DatasetMemberships(client=client)
        self.datasets = Datasets(client=client)
        self.annotations = Annotations(client=client)


class AsyncDB(AsyncAPIResource):
    """Async DB API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.files = AsyncFiles(client=client)
        self.dataset_memberships = AsyncDatasetMemberships(client=client)
        self.datasets = AsyncDatasets(client=client)
        self.annotations = AsyncAnnotations(client=client)
