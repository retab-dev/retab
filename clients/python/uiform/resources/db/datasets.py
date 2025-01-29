from typing import Any, Optional, Literal, List, Dict
from pathlib import Path
from tqdm import tqdm
import json

from ..._resource import SyncAPIResource, AsyncAPIResource

from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.modalities import Modality
from ...types.db.datasets import Dataset, DatasetAnnotationStatus
from ...types.db.dataset_memberships import Annotation




class Datasets(SyncAPIResource):
    """Datasets API wrapper"""



    def create(
        self,
        name: str,
        json_schema: Optional[Dict[str, Any]],
        description: str | None = None
    ) -> Dataset:
        """Create a new dataset.
        
        Args:
            json_schema: The JSON schema for annotations
            name: Name of the dataset
            description: Optional description of the dataset
            
        Returns:
            Dataset: The created dataset object
        """
        data = {
            "json_schema": json_schema,
            "name": name,
            "description": description
        }
        response = self._client._request("POST", "/api/v1/db/datasets", data=data, raise_for_status=True)
        return Dataset(**response)
    
    def clone(
        self,
        dataset_id: str,
        name: str,
        description: str | None = None,
        json_schema: Dict[str, Any] | None = None
    ) -> Dataset:
        """Clone an existing dataset.
        
        Args:
            dataset_id: The ID of the dataset to clone
            name: Name for the cloned dataset
            description: Optional description for the cloned dataset
            json_schema: Optional JSON schema for the cloned dataset. If not provided, uses the original schema.
            
        Returns:
            Dataset: The cloned dataset object
        """
        data = {
            "dataset_id": dataset_id,
            "name": name,
            "description": description,
            "json_schema": json_schema
        }
        response = self._client._request("POST", f"/api/v1/db/datasets/clone", data=data)
        return Dataset(**response)

    def get(self, dataset_id: str) -> Dataset:
        """Get a specific dataset.
        
        Args:
            dataset_id: The ID of the dataset to retrieve
            
        Returns:
            Dataset: The dataset object
        """
        response = self._client._request("GET", f"/api/v1/db/datasets/{dataset_id}")
        return Dataset(**response)

    def list(
        self,
        schema_version: str | None = None,
        schema_data_version: str | None = None,
        dataset_id: str | None = None,
        dataset_name: str | None = None,
        after: str | None = None,
        before: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc"
    ) -> List[Dataset]:
        """List datasets with pagination support.
        
        Args:
            schema_version: Optional filter by schema version
            schema_data_version: Optional filter by schema data version
            dataset_id: Optional filter by dataset ID
            dataset_name: Optional filter by dataset name
            after: An object ID that defines your place in the list
            before: An object ID that defines your place in the list
            limit: Maximum number of datasets to return (1-100)
            order: Sort order by creation time ("asc" or "desc")
            
        Returns:
            List[Dataset]: List of dataset objects
        """
        params: dict[str, str | int] = {"limit": limit}
        if schema_version:
            params["schema_version"] = schema_version
        if schema_data_version:
            params["schema_data_version"] = schema_data_version
        if dataset_id:
            params["dataset_id"] = dataset_id
        if dataset_name:
            params["dataset_name"] = dataset_name
        if after:
            params["after"] = after
        if before:
            params["before"] = before
        if order:
            params["order"] = order
            
        response = self._client._request("GET", "/api/v1/db/datasets", params=params, raise_for_status=True)
        return [Dataset(**item) for item in response["items"]]

    def update(
        self,
        dataset_id: str,
        name: str | None = None,
        description: str | None = None,
        json_schema: Dict[str, Any] | None = None
    ) -> Dataset:
        """Update a dataset's properties. json_schema is not updatable. If you want to update it, you have to delete the dataset and create a new one.
        
        Args:
            dataset_id: The ID of the dataset to update
            name: Optional new name for the dataset
            description: Optional new description
            json_schema: Optional new JSON schema for the dataset
        Returns:
            Dataset: The updated dataset object
        """
        data: dict[str, Any] = {}
        if json_schema is not None:
            data["json_schema"] = json_schema
        if name is not None:
            data["name"] = name
        if description is not None:
            data["description"] = description
            
        response = self._client._request("PUT", f"/api/v1/db/datasets/{dataset_id}", data=data)
        return Dataset(**response)

    def delete(self, dataset_id: str, force: bool = False) -> None:
        """Delete a dataset.
        
        Args:
            dataset_id: The ID of the dataset to delete
            force: If True, delete the dataset even if it has annotations
        """

        self._client._request("DELETE", f"/api/v1/db/datasets/{dataset_id}", params={"force": force})

    def annotation_status(self, dataset_id: str) -> DatasetAnnotationStatus:
        """Get annotation statistics for a dataset.

        Args:
            dataset_id: The ID of the dataset to get stats for

        Returns:
            Dict containing:
                - total_files: Total number of files in dataset
                - files_without_annotations: Number of files without annotations
                - status_counts: Count of each annotation status
                - files_without_annotation_ids: List of file IDs without annotations
        """
        response = self._client._request(
            "GET",
            f"/api/v1/db/datasets/{dataset_id}/annotation-status"
        )
        return DatasetAnnotationStatus(**response)

    def download(
        self,
        dataset_id: str,
        path: Path | str,
        format: Literal["jsonl", "zip"] = "jsonl",
        text_operations: Optional[dict[str, Any]] = None,
        image_operations: Optional[dict[str, Any]] = None,
        modality: Modality = "native",
    ) -> None:
        """Download a dataset's contents to a file.
        
        Args:
            dataset_id: The ID of the dataset to download
            path: Path where to save the dataset
            format: Output format, either "jsonl" for training data or "zip" for raw files
        
        Raises:
            ValueError: If any files are missing annotations or have incomplete/empty status
        """
        if format == "zip":
            raise NotImplementedError("ZIP format not yet supported")
        
        # Check annotation status
        status = self.annotation_status(dataset_id)
        if len(status.files_with_empty_annotations) > 0 or len(status.files_with_incomplete_annotations) > 0:
            raise ValueError(f"Dataset has files with missing or incomplete annotations: Files with empty annotations={len(status.files_with_empty_annotations)}, Files with incomplete annotations={len(status.files_with_incomplete_annotations)}, Files with completed annotations={len(status.files_with_completed_annotations)}")
        
        # Get all files and annotations in dataset
        annotations = self._client.db.annotations.list(dataset_id=dataset_id, limit=1000000)
        # Get all dataset memberships
        memberships = self._client.db.dataset_memberships.list(dataset_id=dataset_id)
        
        # Create lookup of file_id -> annotation
        file_ids = {m.file_id: m.annotation for m in memberships}
        
        with open(path, 'w', encoding='utf-8') as f:
            for file_id in tqdm(file_ids, desc="Processing files", position=0):
                # Get file download link
                file_link = self._client.db.files.download_link(file_id)
                # Create document message
                msg_obj = self._client.documents.create_messages(
                    document=file_link.download_url,
                    modality=modality,
                    text_operations=text_operations,
                    image_operations=image_operations
                )
                
                # Get corresponding annotation
                annotation = file_ids.get(file_id)
                assert isinstance(annotation, Annotation)

                assistant_message: List[ChatCompletionUiformMessage] = [
                    {"role": "assistant", "content": json.dumps(annotation.data, ensure_ascii=False, indent=2)}
                ]
                
                # Write entry in JSONL format
                entry = {
                    "messages": msg_obj.messages + assistant_message
                }
                f.write(json.dumps(entry) + '\n')


    def annotate(
        self,
        dataset_id: str,
        model: str = "gpt-4",
        modality: Modality = "native"
    ) -> None:
        """Start an annotation job for a dataset.
        
        Args:
            dataset_id: The ID of the dataset to annotate
            model: The model to use for annotation
            modality: The modality to use for processing
        """
        data = {
            "model": model,
            "modality": modality
        }
        raise NotImplementedError("Annotate is not yet implemented")
    
"""

    def bulk_annotate(
        self,
        annotated_dataset_id: str,
        annotations_file: IOBase | Path | str
    ) -> BulkAnnotationResponse:
        #Bulk annotate files -> Still remains to be implemented
        files = self._prepare_file_upload([annotations_file])
        response = self._client._request(
            "POST",
            f"/api/v1/datasets/annotated-datasets/{annotated_dataset_id}/bulk-annotate",
            files=[("files", files[0])]
        )
        return BulkAnnotationResponse.model_validate(response)

        """
