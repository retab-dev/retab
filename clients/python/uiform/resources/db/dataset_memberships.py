from typing import Any, Optional, Literal, List, Dict

from ..._resource import SyncAPIResource, AsyncAPIResource

from ...types.db.dataset_memberships import DatasetMembership




class DatasetMemberships(SyncAPIResource):
    """Dataset Memberships API wrapper for managing file associations with datasets."""

    def create(self, file_id: str, dataset_id: str) -> DatasetMembership:
        """Create a new dataset membership.
        
        Args:
            file_id: The ID of the file to add to the dataset
            dataset_id: The ID of the dataset to add the file to
            
        Returns:
            DatasetMembership: The created dataset membership object
        """
        data = {
            "file_id": file_id,
            "dataset_id": dataset_id
        }
        response = self._client._request("POST", "/v1/db/dataset-memberships", data=data)
        return DatasetMembership(**response)

    def get(self, dataset_dataset_membership_id: str) -> DatasetMembership:
        """Get a specific dataset membership.
        
        Args:
            dataset_dataset_membership_id: The ID of the dataset membership to retrieve
            
        Returns:
            DatasetMembership: The dataset membership object
        """
        response = self._client._request("GET", f"/v1/db/dataset-memberships/{dataset_dataset_membership_id}")
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
            
        response = self._client._request("GET", "/v1/db/dataset-memberships", params=params)
        return [DatasetMembership(**item) for item in response["items"]]

    def delete(self, dataset_dataset_membership_id: str) -> None:
        """Delete a dataset membership.
        
        Args:
            dataset_dataset_membership_id: The ID of the dataset membership to delete
        """
        self._client._request("DELETE", f"/v1/db/dataset-memberships/{dataset_dataset_membership_id}")


