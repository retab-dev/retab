import datetime
from typing import Any, Optional

import nanoid  # type: ignore
from pydantic import AliasChoices, BaseModel, ConfigDict, Field, computed_field, field_validator

from ..documents.split import Subdocument
from ..inference_settings import InferenceSettings
from ..mime import BaseMIMEData, MIMEData
from ..projects.predictions import PredictionData
from ..projects.model import default_inference_settings


class SplitDraftConfig(BaseModel):
    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    inference_settings: InferenceSettings = Field(default=default_inference_settings)
    split_config: list[Subdocument] = Field(
        default_factory=list,
        validation_alias=AliasChoices("split_config", "subdocuments"),
        serialization_alias="split_config",
    )
    json_schema: dict[str, Any] = Field(default_factory=dict)
    subdocuments: list[Subdocument] = Field(default_factory=list)


class SplitPublishedConfig(SplitDraftConfig):
    origin: str = Field(default="manual")


class SplitProject(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str = Field(default_factory=lambda: "project_" + nanoid.generate())
    name: str = Field(default="")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    published_config: SplitPublishedConfig = Field(default_factory=SplitPublishedConfig)
    draft_config: SplitDraftConfig = Field(default_factory=SplitDraftConfig)
    is_published: bool = False


class CreateSplitProjectRequest(BaseModel):
    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    name: str
    split_config: list[Subdocument] = Field(
        default_factory=list,
        validation_alias=AliasChoices("split_config", "subdocuments"),
        serialization_alias="split_config",
    )


class PatchSplitProjectRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    name: Optional[str] = Field(default=None)
    published_config: Optional[SplitPublishedConfig] = Field(default=None)
    draft_config: Optional[SplitDraftConfig] = Field(default=None)
    is_published: Optional[bool] = Field(default=None)


class SplitDataset(BaseModel):
    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    id: str
    name: str = Field(default="", description="The name of the dataset")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    base_split_config: list[Subdocument] = Field(
        default_factory=list,
        validation_alias=AliasChoices("base_split_config", "base_json_schema", "subdocuments", "base_subdocuments"),
        serialization_alias="base_split_config",
    )
    base_json_schema: dict[str, Any] = Field(default_factory=dict)
    base_subdocuments: list[Subdocument] = Field(default_factory=list)
    base_inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)
    project_id: str


class CreateSplitDatasetRequest(BaseModel):
    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    name: str
    base_split_config: list[Subdocument] = Field(
        default_factory=list,
        validation_alias=AliasChoices("base_split_config", "base_json_schema", "subdocuments", "base_subdocuments"),
        serialization_alias="base_split_config",
    )
    base_inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)


class PatchSplitDatasetRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    name: Optional[str] = Field(default=None, description="The name of the dataset")


class SplitDatasetDocument(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    project_id: str
    dataset_id: str
    mime_data: BaseMIMEData = Field(description="The mime data of the document")
    prediction_data: PredictionData = Field(default_factory=PredictionData)
    extraction_id: str | None = Field(default=None)
    split_id: str | None = Field(default=None)
    validation_flags: dict[str, Any] = Field(default_factory=dict)


class SplitBuilderDocument(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    project_id: str
    mime_data: BaseMIMEData = Field(description="The mime data of the document")
    prediction_data: PredictionData = Field(default_factory=PredictionData)
    split_id: str | None = Field(default=None)
    extraction_id: str | None = Field(default=None)


class PatchSplitDatasetDocumentRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    validation_flags: Optional[dict[str, Any]] = Field(default=None)
    prediction_data: Optional[PredictionData] = Field(default=None)
    extraction_id: Optional[str] = Field(default=None)


class SplitConfigOverrides(BaseModel):
    model_config = ConfigDict(extra="ignore", populate_by_name=True)

    descriptions_override: Optional[dict[str, str]] = Field(
        default=None,
        validation_alias=AliasChoices("descriptions_override", "descriptionsOverride"),
        serialization_alias="descriptions_override",
    )
    partition_keys_override: Optional[dict[str, str | None]] = Field(
        default=None,
        validation_alias=AliasChoices("partition_keys_override", "partitionKeysOverride"),
        serialization_alias="partition_keys_override",
    )


class SplitDraftIteration(BaseModel):
    model_config = ConfigDict(extra="ignore")

    split_config_overrides: SplitConfigOverrides = Field(
        default_factory=SplitConfigOverrides,
        validation_alias=AliasChoices("split_config_overrides", "config_overrides", "schema_overrides"),
        serialization_alias="split_config_overrides",
    )
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)


class SplitIteration(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)
    split_config_overrides: SplitConfigOverrides = Field(
        default_factory=SplitConfigOverrides,
        validation_alias=AliasChoices("split_config_overrides", "config_overrides", "schema_overrides"),
        serialization_alias="split_config_overrides",
    )
    parent_id: Optional[str] = Field(default=None)
    project_id: str
    dataset_id: str
    draft: SplitDraftIteration = Field(default_factory=SplitDraftIteration)
    status: str = Field(default="draft")
    finalized_at: datetime.datetime | None = Field(default=None)
    last_finalize_error: str | None = Field(default=None)

    @field_validator("status", mode="before")
    @classmethod
    def _normalize_status(cls, value: str | None) -> str:
        if value == "completed":
            return "finalized"
        return value or "draft"


class CreateSplitIterationRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    inference_settings: InferenceSettings = Field(default_factory=InferenceSettings)
    split_config_overrides: SplitConfigOverrides = Field(
        default_factory=SplitConfigOverrides,
        validation_alias=AliasChoices("split_config_overrides", "config_overrides", "schema_overrides"),
        serialization_alias="split_config_overrides",
    )
    project_id: str
    dataset_id: str
    parent_id: Optional[str] = Field(default=None)


class PatchSplitIterationRequest(BaseModel):
    model_config = ConfigDict(extra="ignore")

    inference_settings: Optional[InferenceSettings] = Field(default=None)
    split_config_overrides: Optional[SplitConfigOverrides] = Field(
        default=None,
        validation_alias=AliasChoices("split_config_overrides", "config_overrides", "schema_overrides"),
        serialization_alias="split_config_overrides",
    )
    draft: Optional[SplitDraftIteration] = Field(default=None)


class SplitIterationDocument(BaseModel):
    model_config = ConfigDict(extra="ignore")

    id: str
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(tz=datetime.timezone.utc))
    project_id: str
    iteration_id: str
    dataset_id: str
    dataset_document_id: str
    mime_data: BaseMIMEData = Field(description="The mime data of the document")
    prediction_data: PredictionData = Field(default_factory=PredictionData)
    extraction_id: str | None = Field(default=None)
    split_id: str | None = Field(default=None)
