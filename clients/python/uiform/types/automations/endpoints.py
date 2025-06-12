import nanoid  # type: ignore
from pydantic import BaseModel, Field, computed_field

from ..logs import AutomationConfig, UpdateAutomationRequest
from ..pagination import ListMetadata


class Endpoint(AutomationConfig):
    @computed_field
    @property
    def object(self) -> str:
        return "automation.endpoint"

    id: str = Field(default_factory=lambda: "endp_" + nanoid.generate(), description="Unique identifier for the extraction endpoint")


class ListEndpoints(BaseModel):
    data: list[Endpoint]
    list_metadata: ListMetadata


# Inherits from the methods of UpdateAutomationRequest
class UpdateEndpointRequest(UpdateAutomationRequest):
    pass
