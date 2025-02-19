from pydantic import BaseModel, Field, field_serializer
from typing import Literal, Any
import uuid
import datetime
from ..image_settings import ImageSettings
from ..modalities import Modality
from pydantic import HttpUrl
from pydantic_core import Url



class ExtractionEndpointConfig(BaseModel):
    object: Literal['endpoint'] = "endpoint"
    id: str = Field(default_factory=lambda: "endp" + str(uuid.uuid4()), description="Unique identifier for the extraction endpoint")
    
    # Extraction Endpoint Specific Config
    name: str = Field(..., description="Name of the extraction endpoint")

    # Automation Config
    webhook_url: HttpUrl = Field(..., description = "Url of the webhook to send the data to")
    webhook_headers: dict[str, str] = Field(default_factory=dict, description = "Headers to send with the request")
    updated_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))
    # DocumentExtraction Config
    modality: Modality
    image_settings : ImageSettings = Field(default_factory=ImageSettings, description="Preprocessing operations applied to image before sending them to the llm")
    model: str = Field(..., description="Model used for chat completion")
    json_schema: dict[str, Any] = Field(..., description="JSON schema format used to validate the output data.")
    temperature: float = Field(default=0.0, description="Temperature for sampling. If not provided, the default temperature for the model will be used.", examples=[0.0])


    @field_serializer('webhook_url')
    def url2str(self, val: HttpUrl) -> str:
        return str(val)


