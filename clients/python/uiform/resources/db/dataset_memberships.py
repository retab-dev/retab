from typing import Any, Optional, Literal, List, Dict

from ..._resource import SyncAPIResource, AsyncAPIResource
from ...types.standards import PreparedRequest
from ...types.db.dataset_memberships import DatasetMembership


class DatasetMembershipsMixin:
    def prepare_create(self, file_id: str, dataset_id: str) -> PreparedRequest:
        data = {
            "file_id": file_id,
            "dataset_id": dataset_id
        }
        return PreparedRequest(method="POST", url="/v1/db/dataset-memberships", data=data)

    def prepare_get(self, dataset_dataset_membership_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/db/dataset-memberships/{dataset_dataset_membership_id}")
    
    def prepare_list(self, dataset_id: str | None = None, file_id: str | None = None, after: str | None = None, before: str | None = None, limit: int = 10, order: Literal["asc", "desc"] | None = "desc") -> PreparedRequest:
        params: dict[str, str | int] = {"limit": limit}
        if dataset_id:
            params["dataset_id"] = dataset_id
        if file_id:
            params["file_id"] = file_id
        if after:
            params["after"] = after
        if before:
            params["before"] = before
        if order:
            params["order"] = order
        return PreparedRequest(method="GET", url="/v1/db/dataset-memberships", params=params)
    
    def prepare_delete(self, dataset_dataset_membership_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/db/dataset-memberships/{dataset_dataset_membership_id}")

class DatasetMemberships(SyncAPIResource, DatasetMembershipsMixin):
    """Dataset Memberships API wrapper for managing file associations with datasets."""

    def create(self, file_id: str, dataset_id: str) -> DatasetMembership:
        """Create a new dataset membership.
        
        Args:
            file_id: The ID of the file to add to the dataset
            dataset_id: The ID of the dataset to add the file to
            
        Returns:
            DatasetMembership: The created dataset membership object
        """
        request = self.prepare_create(file_id, dataset_id)
        response = self._client._prepared_request(request)
        return DatasetMembership(**response)

    def get(self, dataset_dataset_membership_id: str) -> DatasetMembership:
        """Get a specific dataset membership.
        
        Args:
            dataset_dataset_membership_id: The ID of the dataset membership to retrieve
            
        Returns:
            DatasetMembership: The dataset membership object
        """
        request = self.prepare_get(dataset_dataset_membership_id)
        response = self._client._prepared_request(request)
        return DatasetMembership(**response)

    def list(
        self,
        dataset_id: str | None = None,
        file_id: str | None = None,
        after: str | None = None,
        before: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc"
    ) -> List[DatasetMembership]:
        """List dataset memberships with pagination.
        
        Args:
            dataset_id: Optional filter by dataset ID
            file_id: Optional filter by file ID
            after: An object ID that defines your place in the list
            before: An object ID that defines your place in the list
            limit: Maximum number of memberships to return (1-100)
            order: Sort order by creation time ("asc" or "desc")
            
        Returns:
            List[DatasetMembership]: List of dataset membership objects
        """
        params: dict[str, str | int] = {"limit": limit}
        if dataset_id:
            params["dataset_id"] = dataset_id
        if file_id:
            params["file_id"] = file_id
        if after:
            params["after"] = after
        if before:
            params["before"] = before
        if order:
            params["order"] = order
            
        request = self.prepare_list(dataset_id, file_id, after, before, limit, order)
        response = self._client._prepared_request(request)
        return [DatasetMembership(**item) for item in response["items"]]

    def delete(self, dataset_dataset_membership_id: str) -> None:
        """Delete a dataset membership.
        
        Args:
            dataset_dataset_membership_id: The ID of the dataset membership to delete
        """
        request = self.prepare_delete(dataset_dataset_membership_id)
        self._client._prepared_request(request)


class AsyncDatasetMemberships(AsyncAPIResource, DatasetMembershipsMixin):
    async def create(self, file_id: str, dataset_id: str) -> DatasetMembership:
        request = self.prepare_create(file_id, dataset_id)
        response = await self._client._prepared_request(request)
        return DatasetMembership(**response)

    async def get(self, dataset_dataset_membership_id: str) -> DatasetMembership:
        request = self.prepare_get(dataset_dataset_membership_id)
        response = await self._client._prepared_request(request)
        return DatasetMembership(**response)
    
    async def list(self, dataset_id: str | None = None, file_id: str | None = None, after: str | None = None, before: str | None = None, limit: int = 10, order: Literal["asc", "desc"] | None = "desc") -> List[DatasetMembership]:
        request = self.prepare_list(dataset_id, file_id, after, before, limit, order)
        response = await self._client._prepared_request(request)
        return [DatasetMembership(**item) for item in response["items"]]
    
    async def delete(self, dataset_dataset_membership_id: str) -> None:
        request = self.prepare_delete(dataset_dataset_membership_id)
        await self._client._prepared_request(request)
