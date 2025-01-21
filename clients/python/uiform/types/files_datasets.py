from pydantic import BaseModel, Field, HttpUrl
from typing import List, Optional, Any, Literal, Tuple, BinaryIO
from datetime import datetime

FileData = Tuple[str, BinaryIO, str]
# Type for (field_name, (filename, file, content_type))
FileTuple = Tuple[str, FileData]

# Base types for files and datasets
class DatasetFile(BaseModel):
    id: str = Field(description="The unique identifier of the file")
    filename: str = Field(description="The name of the file")
    mime_type: str = Field(description="The MIME type of the file (e.g., 'application/pdf', 'image/png')")
    gcs_path: str = Field(description="The path in the GCS bucket")
    size_in_bytes: Optional[int] = Field(default=None, description="The size of the file in bytes")
    uploaded_at: Optional[datetime] = Field(default=None, description="Timestamp for when the file was uploaded")
    modified_at: Optional[datetime] = Field(default=None, description="Timestamp for the last modification")
    user_id: Optional[str] = Field(default=None, description="ID of the user associated with the file")
    organization_id: str = Field(description="ID of the organization associated with the file")
    datasets_ids: List[str] = Field(default_factory=list, description="List of dataset IDs the file belongs to")

class Dataset(BaseModel):
    id: str = Field(description="Unique identifier of the dataset")
    name: str = Field(description="Name of the dataset")
    description: Optional[str] = Field(default=None, description="Optional description of the dataset")
    created_at: datetime = Field(description="Timestamp for when the dataset was created")
    updated_at: Optional[datetime] = Field(default=None, description="Timestamp for when the dataset was last updated")
    organization_id: str = Field(description="ID of the organization associated with the dataset")

# Annotation related types
AnnotationStatus = Literal["not-annotated", "in-progress", "annotated"]

class AnnotatedFile(BaseModel):
    id: str = Field(description="Unique identifier of the annotated file")
    file_id: str = Field(description="ID of the file being annotated")
    annotation: dict[str, Any] = Field(description="The annotation of the file")
    user_id: str = Field(description="ID of the user who annotated the file")
    annotated_at: datetime = Field(description="Timestamp for when the file was annotated")
    annotated_dataset_id: str = Field(description="ID of the annotated dataset")
    status: AnnotationStatus = Field(
        default="not-annotated",
        description="Status of the annotation"
    )

class AnnotatedDataset(BaseModel):
    id: str = Field(description="Unique identifier of the annotated dataset")
    dataset_id: str = Field(description="ID of the dataset being annotated")
    json_schema: dict[str, Any] = Field(description="The JSON schema of the annotated dataset")
    description: Optional[str] = Field(default=None, description="Optional description of the annotated dataset")
    created_at: datetime = Field(description="Timestamp for when the annotated dataset was created")

# Request models
class CreateDatasetRequest(BaseModel):
    name: str = Field(description="Name of the dataset")
    description: Optional[str] = Field(default=None, description="Optional description of the dataset")

class UpdateAnnotatedDatasetRequest(BaseModel):
    description: Optional[str] = Field(default=None, description="Updated description for the annotated dataset")

class UpdateAnnotationRequest(BaseModel):
    annotation: Optional[dict[str, Any]] = Field(default=None, description="Updated annotation data")
    status: Optional[AnnotationStatus] = Field(default=None, description="Updated annotation status")

# Response models
class UploadFileResponse(BaseModel):
    url: str = Field(description="The URL of the uploaded file")
    id: str = Field(description="The ID of the uploaded file")

class MultipleUploadResponse(BaseModel):
    files: List[UploadFileResponse] = Field(description="List of uploaded files")

class ListDatasetsResponse(BaseModel):
    datasets: List[Dataset] = Field(description="List of datasets")
    total_count: int = Field(description="Total number of datasets")

class PaginatedListFilesResponse(BaseModel):
    files: List[DatasetFile] = Field(description="List of files")
    total: int = Field(description="Total number of files")
    page: int = Field(description="Current page number")
    page_size: int = Field(description="Number of items per page")
    total_pages: int = Field(description="Total number of pages")

class FileWithAnnotation(BaseModel):
    file: DatasetFile = Field(description="The file details")
    annotation: Optional[AnnotatedFile] = Field(default=None, description="The file's annotation if it exists")

class FileWithAnnotationAndSchema(FileWithAnnotation):
    json_schema: dict[str, Any] = Field(description="The JSON schema of the annotated dataset")

class BulkAnnotationResponse(BaseModel):
    total_processed: int = Field(description="Total number of files processed")
    successful: int = Field(description="Number of successful annotations")
    failed: int = Field(description="Number of failed annotations")
    errors: List[str] = Field(description="List of error messages")

class ListAnnotatedDatasetsResponse(BaseModel):
    annotated_datasets: List[AnnotatedDataset]

class PaginatedListFilesWithAnnotationsResponse(BaseModel):
    files: List[FileWithAnnotation]
    annotated_dataset: AnnotatedDataset
    total: int
    page: int
    page_size: int
    total_pages: int

class FileLink(BaseModel):
    download_url: HttpUrl = Field(description="The signed URL to download the file")
    expires_in: str = Field(description="The expiration time of the signed URL")
    filename: str = Field(description="The name of the file")
