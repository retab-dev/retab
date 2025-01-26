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
