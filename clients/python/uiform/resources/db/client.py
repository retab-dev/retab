from typing import Any

from ..._resource import SyncAPIResource, AsyncAPIResource

from .files import Files
from .dataset_memberships import DatasetMemberships
from .datasets import Datasets
from .annotations import Annotations



class DB(SyncAPIResource): 
    """DB API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.files = Files(client=client)
        self.dataset_memberships = DatasetMemberships(client=client)
        self.datasets = Datasets(client=client)
        self.annotations = Annotations(client=client)

