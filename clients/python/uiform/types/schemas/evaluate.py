from typing import Any, Self

from openai.types.chat.chat_completion_reasoning_effort import ChatCompletionReasoningEffort
from pydantic import BaseModel, Field, model_validator

from ..evals import ItemMetric
from ..mime import MIMEData
from ..modalities import Modality
from ..browser_canvas import BrowserCanvas


class EvaluateSchemaRequest(BaseModel):
    """
    The request body for evaluating a JSON Schema.

    Two evaluation modes are supported:
    - 'consensus': Uses multiple extractions to compute consensus-based likelihoods
    - 'distance': Compares single extraction against provided ground truths
    """

    documents: list[MIMEData]
    ground_truths: list[dict[str, Any]] | None = None
    model: str = "gpt-4o-mini"
    reasoning_effort: ChatCompletionReasoningEffort = "medium"
    modality: Modality
    image_resolution_dpi: int = Field(default=96, description="Resolution of the image sent to the LLM")
    browser_canvas: BrowserCanvas = Field(
        default="A4", description="Sets the size of the browser canvas for rendering documents in browser-based processing. Choose a size that matches the document type."
    )
    n_consensus: int = 1
    json_schema: dict[str, Any]

    @model_validator(mode="after")
    def validate_evaluation_mode_consistency(self) -> Self:
        """Validate that parameters are consistent with the chosen evaluation mode."""

        if self.ground_truths is None:
            # Consensus mode requirements
            if self.n_consensus <= 1:
                raise ValueError("Consensus mode requires n_consensus > 1")
        else:
            # Distance mode requirements
            if self.n_consensus != 1:
                raise ValueError("Distance mode requires n_consensus = 1")
            if len(self.documents) != len(self.ground_truths):
                raise ValueError("Distance mode requires equal number of documents and ground_truths")
        if len(self.documents) == 0:
            raise ValueError("Evaluation mode requires at least one document")

        return self


class EvaluateSchemaResponse(BaseModel):
    """
    The response body for evaluating a JSON Schema.
    """

    item_metrics: list[ItemMetric]
