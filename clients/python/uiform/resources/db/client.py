from pydantic import BaseModel, Field
from typing import Any, Optional, Literal, List, Dict
from io import BytesIO
import datetime
import uuid
import base64

from ...types.files_datasets import FileData, FileTuple
from ..._utils.mime import prepare_mime_document
from ..._resource import SyncAPIResource, AsyncAPIResource


from pathlib import Path
from io import IOBase
import PIL.Image
from ...types.mime import MIMEData

from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.modalities import Modality
from ...types.db.files import DBFile
from ...types.db.dataset_memberships import DatasetMembership
from ...types.db.datasets import Dataset
from ...types.db.annotations import Annotation, GenerateAnnotationRequest

class Files(SyncAPIResource):
    """Files API wrapper"""

    def create_file(self,
                    document: Path | str | IOBase | MIMEData | PIL.Image.Image,
                    dataset_id: str | None = None,
                    dataset_name: str | None = None
    ) -> DBFile:
        """Upload the file to the server. 
        We force the user to specify a dataset membership when the user creates a file, to avoid orphan files.
        But if the dataset is then deleted, the file will still be there, and will have to be deleted separately if that's what the user wants.

        Args:
            document: The file to upload
            dataset_id: The ID of the dataset to add the file to
            dataset_name: The name of the dataset to add the file to
        

        Returns:
            File: The created file object
        """

        if dataset_id is None and dataset_name is None:
            raise ValueError("Either dataset_id or dataset_name must be provided")
        if dataset_id is not None and dataset_name is not None:
            raise ValueError("Only one of dataset_id or dataset_name can be provided")

        mime_document = prepare_mime_document(document)

        content_binary = BytesIO(mime_document.content.encode('utf-8'))
        file_data: FileData = (mime_document.name, content_binary, mime_document.mime_type)
        files: List[FileTuple] = [("file", file_data)]

        # Add dataset information as query parameters
        params = {}
        if dataset_id:
            params["dataset_id"] = dataset_id
        if dataset_name:
            params["dataset_name"] = dataset_name

        response = self._client._request("POST", "/api/v1/db/files", files=files, params=params)

        return DBFile(**response)
    
    def get_file(self, file_id: str) -> DBFile:
        """Get a file by ID.
        
        Args:
            file_id: The ID of the file to retrieve
            
        Returns:
            DBFile: The file object
        """
        response = self._client._request("GET", f"/api/v1/db/files/{file_id}")
        return DBFile(**response)
    
    def download_file(self, file_id: str) -> MIMEData:
        """Download a file's content by ID.
        
        Args:
            file_id: The ID of the file to download
            
        Returns:
            MIMEData: The file content and metadata as a MIMEData object
            
        """
        # First get the file metadata
        file = self.get_file(file_id)
        
        # Then download the content
        response = self._client._request(
            "GET",
            f"/api/v1/db/files/{file_id}/download",
        )
        
        # Create and return MIMEData object
        # Assuming the response contains the base64 encoded content directly
        return MIMEData(
            id=file.id,
            name=file.filename,
            content=response["content"] if isinstance(response, dict) else base64.b64encode(response.content).decode(),
            mime_type=file.mime_type
        )
    
    def delete_file(self, file_id: str) -> None:
        """Delete a file by ID.
        
        Args:
            file_id: The ID of the file to delete
            
        """
        self._client._request("DELETE", f"/api/v1/db/files/{file_id}")

    def list_files(
        self,
        dataset_id: str | None = None,
        mime_type: str | None = None,
        filename: str | None = None,
        file_id: str | None = None,
        after: str | None = None,
        before: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc"
    ) -> List[DBFile]:
        """List files with pagination support.

        Args:
            dataset_id: The ID of the dataset to list files from.
            after: An object ID that defines your place in the list. When provided,
                returns objects after this ID. For example, if you receive 100 objects
                ending with "obj_123", you can pass after="obj_123" to fetch the next batch.
            before: An object ID that defines your place in the list. When provided,
                returns objects before this ID. Similar to 'after' but in reverse.
            limit: Upper limit on the number of objects to return, between 1 and 100.
                Defaults to 10 if not specified.
            order: Sort order by creation time. Use "asc" for oldest first,
                "desc" for newest first. Defaults to "desc" if not specified.

        Returns:
            List[DBFile]: A list of file objects matching the query parameters.
        """
        params: dict[str, str | int] = {"limit": limit}
        if dataset_id:
            params["dataset_id"] = dataset_id
        if mime_type:
            params["mime_type"] = mime_type
        if filename:
            params["filename"] = filename
        if file_id:
            params["file_id"] = file_id
        if after:
            params["after"] = after
        if before:
            params["before"] = before
        if order:
            params["order"] = order
            
        response = self._client._request("GET", "/api/v1/db/files", params=params)
        return [DBFile(**item) for item in response["items"]]



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
        response = self._client._request("POST", "/api/v1/db/dataset-memberships", data=data)
        return DatasetMembership(**response)

    def get(self, dataset_membership_id: str) -> DatasetMembership:
        """Get a specific dataset membership.
        
        Args:
            dataset_membership_id: The ID of the dataset membership to retrieve
            
        Returns:
            DatasetMembership: The dataset membership object
        """
        response = self._client._request("GET", f"/api/v1/db/dataset-memberships/{dataset_membership_id}")
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
            
        response = self._client._request("GET", "/api/v1/db/dataset-memberships", params=params)
        return [DatasetMembership(**item) for item in response["items"]]

    def delete(self, dataset_membership_id: str) -> None:
        """Delete a dataset membership.
        
        Args:
            dataset_membership_id: The ID of the dataset membership to delete
        """
        self._client._request("DELETE", f"/api/v1/db/dataset-memberships/{dataset_membership_id}")

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

    def download(
        self,
        dataset_id: str,
        format: Literal["jsonl", "zip"] = "jsonl"
    ) -> bytes:
        """Download a dataset in the specified format.
        
        Args:
            dataset_id: The ID of the dataset to download
            format: The format to download in ("jsonl" or "zip")
            
        Returns:
            bytes: The dataset content in the requested format
            
        Raises:
            ValueError: If an invalid format is specified
        """
        params = {"format": format}
        response = self._client._request("GET", f"/api/v1/db/datasets/{dataset_id}/download", params=params)
        return response

    def annotate(
        self,
        dataset_id: str,
        model: str = "gpt-4",
        modality: Literal["native"] = "native"
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
        self._client._request("POST", f"/api/v1/db/datasets/{dataset_id}/annotate", data=data)


class Annotations(SyncAPIResource):
    """Annotations API wrapper"""

    def create(
        self,
        file_id: str,
        dataset_id: str,
        data: dict[str, Any],
        status: Literal["empty", "incomplete", "completed"] = "completed"
    ) -> Annotation:
        """Create a new annotation.
        
        Args:
            file_id: The ID of the file to annotate
            dataset_id: The ID of the dataset this annotation belongs to
            data: The annotation data
            status: Initial annotation status
            
        Returns:
            Annotation: The created annotation object
        """
        data = {
            "file_id": file_id,
            "dataset_id": dataset_id,
            "data": data,
            "status": status
        }
        response = self._client._request("POST", "/api/v1/db/annotations", data=data, raise_for_status=True)
        return Annotation(**response)

    def get(self, annotation_id: str) -> Annotation:
        """Get an annotation by ID.
        
        Args:
            annotation_id: The ID of the annotation to retrieve
            
        Returns:
            Annotation: The annotation object
        """
        response = self._client._request("GET", f"/api/v1/db/annotations/{annotation_id}")
        return Annotation(**response)

    def list(
        self,
        annotation_id: str | None = None,
        dataset_id: str | None = None,
        file_id: str | None = None,
        status: Literal["empty", "incomplete", "completed"] | None = None,
        after: str | None = None,
        before: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc"
    ) -> List[Annotation]:
        """List annotations with optional filtering.
        
        Args:
            dataset_id: Filter by dataset ID
            file_id: Filter by file ID
            status: Filter by annotation status
            after: Return objects after this ID
            before: Return objects before this ID
            limit: Maximum number of annotations to return (1-100)
            order: Sort order by creation time ("asc" or "desc")
            
        Returns:
            List[Annotation]: List of annotation objects
        """
        params: dict[str, str | int] = {"limit": limit}
        if annotation_id:
            params["annotation_id"] = annotation_id
        if dataset_id:
            params["dataset_id"] = dataset_id
        if file_id:
            params["file_id"] = file_id
        if status:
            params["status"] = status
        if after:
            params["after"] = after
        if before:
            params["before"] = before
        if order:
            params["order"] = order
            
        response = self._client._request("GET", "/api/v1/db/annotations", params=params, raise_for_status=True)
        return [Annotation(**item) for item in response["items"]]

    def update(
        self,
        annotation_id: str,
        status: Literal["empty", "incomplete", "completed"] | None = None,
        data: dict[str, Any] | None = None
    ) -> Annotation:
        """Update an annotation.
        
        Args:
            annotation_id: The ID of the annotation to update
            status: New annotation status
            data: New annotation data
        Returns:
            Annotation: The updated annotation object
        """
        data = {}
        if status is not None:
            data["status"] = status
        if data is not None:
            data["data"] = data
        response = self._client._request("PUT", f"/api/v1/db/annotations/{annotation_id}", data=data)
        return Annotation(**response)

    def delete(self, annotation_id: str) -> None:
        """Delete an annotation.
        
        Args:
            annotation_id: The ID of the annotation to delete
        """
        self._client._request("DELETE", f"/api/v1/db/annotations/{annotation_id}")
    

    
    def generate(
        self,
        dataset_id: str,
        file_id: str,
        model: str,
        modality: Modality = "native",
        text_operations: Optional[dict[str, Any]] = None,
        image_operations: Optional[dict[str, Any]] = None,
        temperature: float = 0.0,
        messages: List[ChatCompletionUiformMessage] = [],
        upsert: bool = False
    ) -> Annotation:
        """Generate an annotation for a file in a dataset using the specified model.
        
        Args:
            dataset_id: The ID of the dataset
            file_id: The ID of the file to annotate
            model: The AI model to use for annotation
            modality: The modality to use for annotation (currently only "native" is supported)
            text_operations: Optional text preprocessing operations
            image_operations: Optional image preprocessing operations  
            temperature: Model temperature setting (0-1)
            messages: Optional list of messages for context
            
        Returns:
            Annotation: The generated annotation object
        """
        data: dict[str, Any] = {
            "dataset_id": dataset_id,
            "file_id": file_id,
            "model": model,
            "modality": modality,
            "temperature": temperature,
            "messages": messages,
            "upsert": upsert
        }
        if text_operations:
            data["text_operations"] = text_operations
        if image_operations:
            data["image_operations"] = image_operations

        request = GenerateAnnotationRequest.model_validate(data)
        response = self._client._request("POST", "/api/v1/db/annotations/generate", data=request.model_dump())

        return Annotation(**response)



class DB(SyncAPIResource): 
    """DB API wrapper"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.files = Files(client=client)
        self.dataset_memberships = DatasetMemberships(client=client)
        self.datasets = Datasets(client=client)
        self.annotations = Annotations(client=client)


