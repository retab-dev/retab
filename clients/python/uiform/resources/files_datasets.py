from io import IOBase
from pathlib import Path
from typing import Any, Optional, List, BinaryIO, cast
from datetime import datetime
import mimetypes
import json
from uiform.types.documents.create_messages import ChatCompletionUiformMessage
from uiform.types.modalities import Modality
from .._resource import AsyncAPIResource, SyncAPIResource
from ..types.files_datasets import (
     AnnotationStatus, FileLink, FileTuple,FileData,
    Dataset,
    AnnotatedDataset,
    AnnotatedFile,
    CreateDatasetRequest,
    ListDatasetsResponse,
    MultipleUploadResponse,
    PaginatedListFilesResponse,
    UpdateAnnotationRequest,
    BulkAnnotationResponse,
    FileWithAnnotationAndSchema,
    ListAnnotatedDatasetsResponse,
    PaginatedListFilesWithAnnotationsResponse
)

class BaseDatasetsMixin:
    def _dump_training_set(self, training_set: list[dict[str, Any]], jsonl_path: Path | str) -> None:
        with open(jsonl_path, 'w', encoding='utf-8') as file:
            for entry in training_set:
                file.write(json.dumps(entry) + '\n')

    def _prepare_file_upload(
            self,
            files: List[IOBase | Path | str],
    ) -> List[FileData]:
        """
        Prepare files for upload by creating tuples of (filename, file_object, content_type)

        Args:
            files: List of file objects or paths

        Returns:
            List of tuples containing (filename, file_object, content_type)
        """
        prepared_files: List[FileData] = []

        for file in files:
            if isinstance(file, (str, Path)):
                path = Path(file)
                file_obj = cast(BinaryIO, open(path, 'rb'))
                filename = str(path.name)
                content_type = mimetypes.guess_type(path)[0] or "application/octet-stream"
            else:
                file_obj = cast(BinaryIO, file)
                filename = str(getattr(file, 'name', 'file'))
                content_type = str(getattr(file, 'content_type', 'application/octet-stream'))
            file_data: FileData = (filename, file_obj, content_type)
            prepared_files.append(file_data)

        return prepared_files
   

class Datasets(SyncAPIResource, BaseDatasetsMixin):
    """Datasets API wrapper"""

    def create_dataset(
        self,
        name: str,
        description: Optional[str] = None,
    ) -> Dataset:
        """Create a new dataset"""
        request = CreateDatasetRequest(name=name, description=description)
        response = self._client._request(
            "POST",
            "/api/v1/datasets/datasets",
            data=request.model_dump()
        )
        print(response)
        return Dataset.model_validate(response)

    def list_datasets(
        self,
        skip: int = 0,
        limit: int = 10,
        search: Optional[str] = None
    ) -> ListDatasetsResponse:
        """List all datasets"""
        params: dict[str,Any] = {
            "skip": skip,
            "limit": limit
        }
        if search:
            params["search"] = search

        response = self._client._request(
            "GET",
            "/api/v1/datasets/datasets",
            data=params
        )
        return ListDatasetsResponse.model_validate(response)

    def get_dataset(self, dataset_id: str) -> Dataset:
        """Get a specific dataset"""
        response = self._client._request(
            "GET",
            f"/api/v1/datasets/datasets/{dataset_id}"
        )
        return Dataset.model_validate(response)

    def upload_files(
        self,
        files: List[IOBase | Path | str],
        dataset_id: str
    ) -> MultipleUploadResponse:
        """Upload files to a dataset"""
        prepared_files = self._prepare_file_upload(files)
        files_data = [("files", f) for f in prepared_files]
        response = self._client._request(
            "POST",
            "/api/v1/datasets/files",
            files=files_data,
            data={"dataset_id": dataset_id}
        )
        return MultipleUploadResponse.model_validate(response)

    def list_files(
        self,
        dataset_id: Optional[str] = None,
        page: int = 1,
        page_size: int = 10,
        sort_by: Optional[str] = None,
        sort_order: Optional[str] = None
    ) -> PaginatedListFilesResponse:
        """List files in a dataset"""
        params:dict[str,Any] = {
            "page": page,
            "page_size": page_size
        }
        if dataset_id:
            params["dataset_id"] = dataset_id
        if sort_by:
            params["sort_by"] = sort_by
        if sort_order:
            params["sort_order"] = sort_order

        response = self._client._request(
            "GET",
            "/api/v1/datasets/files",
            data=params
        )
        return PaginatedListFilesResponse.model_validate(response)

    def delete_file(self, file_id: str) -> dict:
        """Delete a file"""
        return self._client._request(
            "DELETE",
            f"/api/v1/datasets/files/{file_id}"
        )

    def create_annotated_dataset(
        self,
        dataset_id: str,
        json_schema: dict[str, Any],
        description: Optional[str] = None
    ) -> AnnotatedDataset:
        """Create an annotated dataset"""
        response = self._client._request(
            "POST",
            f"/api/v1/datasets/datasets/{dataset_id}/annotated-datasets",
            data={
                "dataset_id": dataset_id,
                "json_schema": json_schema,
                "description": description
            }
        )
        return AnnotatedDataset.model_validate(response)
    
    def create_annotated_file(
            self,
            annotated_dataset_id: str,
            file_id: str,
            annotation: dict[str, Any],
            status:Optional[AnnotationStatus] = "not-annotated"
    ) -> AnnotatedFile:
        """
        Create an annotated file in an annotated dataset.

        Args:
            annotated_dataset_id: ID of the annotated dataset
            file_id: ID of the file to annotate
            annotation: Annotation data as dictionary

        Returns:
            AnnotatedFile: The created annotated file
        """
        response = self._client._request(
            "POST",
            f"/api/v1/datasets/annotated-datasets/{annotated_dataset_id}/files",
            data={
                "file_id": file_id,
                "annotation": annotation,
                "status":status
            }
        )
        return AnnotatedFile.model_validate(response)

    def get_annotated_file(
        self,
        annotated_dataset_id: str,
        annotated_file_id: str
    ) -> AnnotatedFile:
        """Get an annotated file"""
        response = self._client._request(
            "GET",
            f"/api/v1/datasets/annotated-datasets/{annotated_dataset_id}/files/{annotated_file_id}"
        )
        response_parsed= FileWithAnnotationAndSchema.model_validate(response)
        assert response_parsed.annotation
        return response_parsed.annotation

    def update_annotation(
        self,
        annotated_dataset_id: str,
        annotated_file_id: str,
        annotation: Optional[dict[str, Any]] = None,
        status: Optional[AnnotationStatus] = None
    ) -> AnnotatedFile:
        """Update an annotation"""
        request = UpdateAnnotationRequest(annotation=annotation, status=status)
        response = self._client._request(
            "PUT",
            f"/api/v1/datasets/annotated-datasets/{annotated_dataset_id}/files/{annotated_file_id}",
            data=request.model_dump()
        )
        return AnnotatedFile.model_validate(response)


    def bulk_annotate(
        self,
        annotated_dataset_id: str,
        annotations_file: IOBase | Path | str
    ) -> BulkAnnotationResponse:
        """Bulk annotate files"""
        files = self._prepare_file_upload([annotations_file])
        response = self._client._request(
            "POST",
            f"/api/v1/datasets/annotated-datasets/{annotated_dataset_id}/bulk-annotate",
            files=[("files", files[0])]
        )
        return BulkAnnotationResponse.model_validate(response)

    def list_annotated_datasets(
            self,
            dataset_id: str
    ) -> ListAnnotatedDatasetsResponse:
        """
        List all annotated datasets for a given dataset

        Args:
            dataset_id: ID of the dataset

        Returns:
            ListAnnotatedDatasetsResponse containing the list of annotated datasets
        """
        response = self._client._request(
            "GET",
            f"/api/v1/datasets/datasets/{dataset_id}/annotated-datasets"
        )
        return ListAnnotatedDatasetsResponse.model_validate(response)
    def list_annotated_dataset_files(
        self,
        annotated_dataset_id: str,
        page: int = 1,
        page_size: int = 10,
        sort_by: Optional[str] = None,
        sort_order: Optional[str] = None,
        status: Optional[AnnotationStatus] = None
    ) -> PaginatedListFilesWithAnnotationsResponse:
        """
        List all files in an annotated dataset with their annotations

        Args:
            annotated_dataset_id: ID of the annotated dataset
            page: Page number for pagination
            page_size: Number of items per page
            sort_by: Field to sort by
            sort_order: Sort order ("asc" or "desc")
            status: Filter by annotation status

        Returns:
            PaginatedListFilesWithAnnotationsResponse containing the list of files with annotations
        """
        params: dict[str, Any] = {
            "page": page,
            "page_size": page_size
        }

        if sort_by:
            params["sort_by"] = sort_by
        if sort_order:
            params["sort_order"] = sort_order
        if status:
            params["status"] = status

        response = self._client._request(
            "GET",
            f"/api/v1/datasets/annotated-datasets/{annotated_dataset_id}/files",
            data=params
        )
        return PaginatedListFilesWithAnnotationsResponse.model_validate(response)
    def get_file_download_url(
            self,
            file_path: str
    ) -> FileLink:
        """
        Get a signed URL for accessing a stored object.

        Args:
            file_path: Path of the file in storage

        Returns:
            dict containing:
                - download_url: Signed URL for downloading the file
                - expires_in: Expiration time of the URL
                - filename: Name of the file
        """
        response = self._client._request(
            "GET",
            f"/api/v1/datasets/stored-object/{file_path}"
        )
        return FileLink.model_validate(response)

    def save_annotated_dataset(
            self,
            annotated_dataset_id: str,
            output_path: Path | str,
            text_operations: Optional[dict[str, Any]] = None,
            modality: Modality = "native",
            messages: list[ChatCompletionUiformMessage] = [],
    ) -> None:
        """
       Save all files annotated from an annotated dataset in openai jsonl format

        Args:
            annotated_dataset_id: ID of the annotated dataset
            output_path: output path of jsonl file

        Returns:
            PaginatedListFilesWithAnnotationsResponse containing the list of files with annotations

        """
        page = 1
        total_pages = 1
        training_set = []
        while page <= total_pages:
            # Get batch of annotated files
            response = self.list_annotated_dataset_files(
                annotated_dataset_id=annotated_dataset_id,
                page=page,
                page_size=100,
                status="annotated"
            )

            total_pages = response.total_pages

            # Process each file in the batch
            for file in response.files:
                if file.annotation:  # Check if annotation exists
                    try:
                        # Create message for this file
                        file_url = self.get_file_download_url(file.file.gcs_path)
                        document_messages = self._client.documents.create_messages(
                            document=file_url.download_url,
                            modality=modality,
                            text_operations=text_operations)
                        assistant_message= {"role": "assistant", "content": json.dumps(file.annotation.annotation, ensure_ascii=False, indent=2)}
                        training_set.append({"messages": document_messages.messages + messages + [assistant_message]})
                    except Exception as e:
                        print(f"Error processing file {file.file.id}: {str(e)}")

            page += 1
        self._dump_training_set(training_set, output_path)

class AsyncDatasets(AsyncAPIResource, BaseDatasetsMixin):
    """Async Datasets API wrapper"""


    # Add  async methods following the same pattern...
