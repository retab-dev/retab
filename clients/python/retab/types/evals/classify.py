import datetime
from typing import Optional

import nanoid  # type: ignore
from pydantic import BaseModel, ConfigDict, Field, field_validator

from ..documents.classify import Category
from ..inference_settings import InferenceSettings
from ..mime import BaseMIMEData
from ..projects.predictions import PredictionData
from ..projects.model import default_inference_settings


class ClassifyDraftConfig(BaseModel):
    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    inference_settings: InferenceSettings = Field(default=default_inference_settings)
    categories: list[Category] = Field(default_factory=list)


class ClassifyPublishedConfig(ClassifyDraftConfig):
    origin: str = Field(default="manual")


class ClassifyProject(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str = Field(default_factory=lambda: "project_" + nanoid.generate())
    name: str = Field(default="")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    published_config: ClassifyPublishedConfig = Field(default_factory=ClassifyPublishedConfig)
    draft_config: ClassifyDraftConfig = Field(default_factory=ClassifyDraftConfig)
    is_published: bool = False


class CreateClassifyProjectRequest(BaseModel):
    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    name: str
    categories: list[Category] = Field(default_factory=list)


class PatchClassifyProjectRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    name: Optional[str] = Field(default=None)
    published_config: Optional[ClassifyPublishedConfig] = Field(default=None)
    draft_config: Optional[ClassifyDraftConfig] = Field(default=None)
    is_published: Optional[bool] = Field(default=None)


class ClassifyDataset(BaseModel):
    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    id: str
    name: str = Field(default="", description="The name of the dataset")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    base_categories: list[Category] = Field(default_factory=list)
    base_inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)
    project_id: str


class CreateClassifyDatasetRequest(BaseModel):
    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    name: str
    base_categories: list[Category] = Field(default_factory=list)
    base_inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)


class PatchClassifyDatasetRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    name: Optional[str] = Field(default=None)


class ClassifyDatasetDocument(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    project_id: str
    dataset_id: str
    mime_data: BaseMIMEData = Field(description="The mime data of the document")
    prediction_data: PredictionData = Field(default_factory=PredictionData)
    classification_id: str | None = Field(default=None)
    extraction_id: str | None = Field(default=None)
    validation_flag: bool | None = Field(default=None)


class ClassifyBuilderDocument(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    project_id: str
    mime_data: BaseMIMEData = Field(description="The mime data of the document")
    prediction_data: PredictionData = Field(default_factory=PredictionData)
    classification_id: str | None = Field(default=None)
    extraction_id: str | None = Field(default=None)


class PatchClassifyDatasetDocumentRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    validation_flag: Optional[bool] = Field(default=None)
    prediction_data: Optional[PredictionData] = Field(default=None)
    classification_id: Optional[str] = Field(default=None)
    extraction_id: Optional[str] = Field(default=None)


class CategoryOverrides(BaseModel):
    model_config = ConfigDict(extra="ignore")

    description_overrides: dict[str, str] = Field(default_factory=dict)


class ClassifyDraftIteration(BaseModel):
    model_config = ConfigDict(extra="ignore")

    category_overrides: CategoryOverrides = Field(default_factory=CategoryOverrides)
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)


class ClassifyIteration(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)
    category_overrides: CategoryOverrides = Field(default_factory=CategoryOverrides)
    parent_id: Optional[str] = Field(default=None)
    project_id: str
    dataset_id: str
    draft: ClassifyDraftIteration = Field(default_factory=ClassifyDraftIteration)
    status: str = Field(default="draft")
    finalized_at: datetime.datetime | None = Field(default=None)
    last_finalize_error: str | None = Field(default=None)

    @field_validator("status", mode="before")
    @classmethod
    def _normalize_status(cls, value: str | None) -> str:
        if value == "completed":
            return "finalized"
        return value or "draft"


class CreateClassifyIterationRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)
    category_overrides: CategoryOverrides = Field(default_factory=CategoryOverrides)
    project_id: str
    dataset_id: str
    parent_id: Optional[str] = Field(default=None)


class PatchClassifyIterationRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    inference_settings: Optional[InferenceSettings] = Field(default=None)
    category_overrides: Optional[CategoryOverrides] = Field(default=None)
    draft: Optional[ClassifyDraftIteration] = Field(default=None)


class ClassifyIterationDocument(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    project_id: str
    iteration_id: str
    dataset_id: str
    dataset_document_id: str
    mime_data: BaseMIMEData = Field(description="The mime data of the document")
    prediction_data: PredictionData = Field(default_factory=PredictionData)
    classification_id: str | None = Field(default=None)
    extraction_id: str | None = Field(default=None)
