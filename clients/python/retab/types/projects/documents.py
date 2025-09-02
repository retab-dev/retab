from typing import Any, Optional

from pydantic import BaseModel, Field, ConfigDict

from ..mime import MIMEData
from .predictions import PredictionMetadata


class AnnotatedDocument(BaseModel):
    model_config = ConfigDict(extra="ignore")
    mime_data: MIMEData = Field(
        description="The mime data of the document. Can also be a BaseMIMEData, which is why we have this id field (to be able to identify the file, but id is equal to mime_data.id)"
    )
    annotation: dict[str, Any] = Field(default={}, description="The ground truth of the document")


class DocumentItem(AnnotatedDocument):
    model_config = ConfigDict(extra="ignore")
    annotation_metadata: Optional[PredictionMetadata] = Field(default=None, description="The metadata of the annotation when the annotation is a prediction")
    playground_extraction: dict[str, Any] = Field(default={}, description="The playground extraction of the document")
    playground_extraction_metadata: Optional[PredictionMetadata] = Field(default=None, description="The metadata of the playground extraction")


class ProjectDocument(DocumentItem):
    model_config = ConfigDict(extra="ignore")
    id: str = Field(description="The ID of the document. Equal to mime_data.id but robust to the case where mime_data is a BaseMIMEData")
    ocr_file_id: Optional[str] = Field(default=None, description="The ID of the OCR file")

class CreateProjectDocumentRequest(DocumentItem):
    model_config = ConfigDict(extra="ignore")


class PatchProjectDocumentRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")
    annotation: Optional[dict[str, Any]] = Field(default=None, description="The ground truth of the document")
    annotation_metadata: Optional[PredictionMetadata] = Field(default=None, description="The metadata of the annotation when the annotation is a prediction")
    ocr_file_id: Optional[str] = Field(default=None, description="The ID of the OCR file")
    playground_extraction: Optional[dict[str, Any]] = Field(default=None, description="The playground extraction of the document")
    playground_extraction_metadata: Optional[PredictionMetadata] = Field(default=None, description="The metadata of the playground extraction")