from pydantic import BaseModel, Field
from typing import Any, Literal
from ..mime import MIMEData
from ..modalities import Modality
from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from ..evals import ItemMetric


class EvaluateSchemaRequest(BaseModel):
    """
    The request body for evaluating a JSON Schema.
    """

    documents: list[MIMEData]
    ground_truths: list[dict[str, Any]] | None = None
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    reasoning_effort: ChatCompletionReasoningEffort = "medium"
    modality: Modality
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: Literal['A3', 'A4', 'A5'] = Field(default='A4', description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type.")
    n_consensus: int = 1
    json_schema: dict[str, Any]


class EvaluateSchemaResponse(BaseModel):
    """
    The response body for evaluating a JSON Schema.
    """

    item_metrics: list[ItemMetric]
