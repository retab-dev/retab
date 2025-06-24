from typing import Optional

from pydantic import BaseModel

from ..inference_settings import InferenceSettings


class AnnotationInputData(BaseModel):
    dataset_id: str
    files_ids: Optional[list[str]] = None
    upsert: bool = False
    inference_settings: InferenceSettings
