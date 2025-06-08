import datetime
from typing import Any, Literal, Optional

import nanoid  # type: ignore
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, HttpUrl, field_serializer
from pydantic_core import Url

from ..modalities import Modality


def scrapping_action(link: HttpUrl) -> dict[str, Any]:
    raise NotImplementedError("Scrapping action not implemented")


class CronSchedule(BaseModel):
    second: Optional[int] = Field(0, ge=0, le=59, description="Second (0-59), defaults to 0")
    minute: int = Field(..., ge=0, le=59, description="Minute (0-59)")
    hour: int = Field(..., ge=0, le=23, description="Hour (0-23)")
    day_of_month: Optional[int] = Field(None, ge=1, le=31, description="Day of the month (1-31), None means any day")
    month: Optional[int] = Field(None, ge=1, le=12, description="Month (1-12), None means every month")
    day_of_week: Optional[int] = Field(None, ge=0, le=6, description="Day of the week (0-6, Sunday = 0), None means any day")

    def to_cron_string(self) -> str:
        return f"{self.second or '*'} {self.minute} {self.hour} {self.day_of_month or '*'} {self.month or '*'} {self.day_of_week or '*'}"


from ..logs import AutomationConfig


class ScrappingConfig(AutomationConfig):
    object: Literal['deployment.scrapping_cron'] = "deployment.scrapping_cron"
    id: str = Field(default_factory=lambda: "scrapping_" + nanoid.generate(), description="Unique identifier for the scrapping job")

    # Scrapping Specific Config
    link: HttpUrl = Field(..., description="Link to be scrapped")
    schedule: CronSchedule

    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))

    # HTTP Config
    webhook_url: HttpUrl = Field(..., description="Url of the webhook to send the data to")
    webhook_headers: dict[str, str] = Field(default_factory=dict, description="Headers to send with the request")

    modality: Modality
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: Literal['A3', 'A4', 'A5'] = Field(default='A4', description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type.")

    # New attributes
    model: str = Field(..., description="Model used for chat completion")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])
    reasoning_effort: ChatCompletionReasoningEffort = Field(
        default="medium", description="The effort level for the model to reason about the input data. If not provided, the default reasoning effort for the model will be used."
    )

    @field_serializer('webhook_url', 'link')
    def url2str(self, val: HttpUrl) -> str:
        return str(val)
