from typing import Any, Optional, Literal, List, Dict

from ..._resource import SyncAPIResource, AsyncAPIResource

from ...types.documents.create_messages import ChatCompletionUiformMessage
from ...types.modalities import Modality
from ...types.db.dataset_memberships import Annotation, GenerateAnnotationRequest



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
