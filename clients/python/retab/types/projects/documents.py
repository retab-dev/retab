from typing import Any, Optional

from pydantic import BaseModel, Field

from ..mime import MIMEData
from ..predictions import PredictionMetadata


class AnnotatedDocument(BaseModel):
    mime_data: MIMEData = Field(
        description="The mime data of the document. Can also be a BaseMIMEData, which is why we have this id field (to be able to identify the file, but id is equal to mime_data.id)"
    )
    annotation: dict[str, Any] = Field(default={}, description="The ground truth of the document")


class DocumentItem(AnnotatedDocument):
    annotation_metadata: Optional[PredictionMetadata] = Field(default=None, description="The metadata of the annotation when the annotation is a prediction")


class ProjectDocument(DocumentItem):
    id: str = Field(description="The ID of the document. Equal to mime_data.id but robust to the case where mime_data is a BaseMIMEData")
    ocr_file_id: Optional[str] = Field(default=None, description="The ID of the OCR file")


class CreateProjectDocumentRequest(DocumentItem):
    pass


class PatchProjectDocumentRequest(BaseModel):
    annotation: Optional[dict[str, Any]] = Field(default=None, description="The ground truth of the document")
    annotation_metadata: Optional[PredictionMetadata] = Field(default=None, description="The metadata of the annotation when the annotation is a prediction")
    ocr_file_id: Optional[str] = Field(default=None, description="The ID of the OCR file")
