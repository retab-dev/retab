from pydantic import BaseModel, Field
from typing import Optional, Literal

CheckPoint = Literal[None, "webcrawl_started", "webcrawl_completed", "webcrawl_failed"]

class WebcrawlInputData(BaseModel):
    url: str
    limit: int = Field(default=3, ge=1, le=100)

class WebcrawlJob(BaseModel):
    job_type: Literal["webcrawl"] = "webcrawl"
    input_data: WebcrawlInputData
    checkpoint: CheckPoint = None
    checkpoint_data: Optional[dict] = None

