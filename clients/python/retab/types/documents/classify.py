from typing import Any

from pydantic import BaseModel, Field, model_validator
from ..mime import MIMEData
from .usage import RetabUsage

class Category(BaseModel):
    name: str = Field(..., description="The name of the category")
    description: str = Field(default="", description="The description of the category")

class ClassifyRequest(BaseModel):
    document: MIMEData = Field(..., description="The document to classify")
    categories: list[Category] = Field(..., description="The categories to classify the document into")
    model: str = Field(default="retab-small", description="The model to use for classification")
    first_n_pages: int | None = Field(default=None, description="Only use the first N pages of the document for classification. Useful for large documents where classification can be determined from early pages.")
    context: str | None = Field(default=None, description="Additional context for classification (e.g., iteration context from a loop)")
    n_consensus: int = Field(default=1, ge=1, le=8, description="Number of classification runs to use for consensus voting. Uses deterministic single-pass when set to 1.")
    bust_cache: bool = Field(default=False, description="If true, skip the LLM cache and force a fresh completion")


class ClassifyResult(BaseModel):
    @model_validator(mode="before")
    @classmethod
    def _coerce_legacy_classification_key(cls, value: Any) -> Any:
        if not isinstance(value, dict):
            return value
        if "category" in value:
            return value
        legacy_category = value.get("classification")
        if legacy_category is None:
            return value
        normalized = dict(value)
        normalized["category"] = legacy_category
        return normalized

    reasoning: str = Field(..., description="The reasoning for the classification decision")
    category: str = Field(..., description="The category name that the document belongs to")

    @property
    def classification(self) -> str:
        return self.category


class ClassifyVote(BaseModel):
    """A single LLM vote from a consensus classification run."""

    @model_validator(mode="before")
    @classmethod
    def _coerce_legacy_classification_key(cls, value: Any) -> Any:
        if not isinstance(value, dict):
            return value
        if "category" in value:
            return value
        legacy_category = value.get("classification")
        if legacy_category is None:
            return value
        normalized = dict(value)
        normalized["category"] = legacy_category
        return normalized

    reasoning: str = Field(..., description="The reasoning produced by this classification vote")
    category: str = Field(..., description="The category chosen by this classification vote")

    @property
    def classification(self) -> str:
        return self.category


class ClassifyChoice(BaseModel):
    classification: ClassifyVote = Field(..., description="One alternative classification vote output")

    @model_validator(mode="before")
    @classmethod
    def _wrap_bare_classification(cls, value: Any) -> Any:
        if not isinstance(value, dict):
            return value
        if "classification" in value and isinstance(value.get("classification"), dict):
            return value
        if "reasoning" in value and ("category" in value or "classification" in value):
            return {"classification": value}
        return value


class ClassifyConsensus(BaseModel):
    choices: list[ClassifyChoice] = Field(
        default_factory=list,
        description="Alternative classification vote outputs used to build the consolidated result",
    )
    likelihood: float | None = Field(
        default=None,
        description="Consensus likelihood score (0.0-1.0) of the winning classification.",
    )


def _build_native_classify_envelope(value: dict[str, Any]) -> dict[str, Any]:
    raw_classification = value.get("classification")
    if not isinstance(raw_classification, dict):
        raw_classification = value.get("result")
    if not isinstance(raw_classification, dict):
        return value

    classification = ClassifyResult.model_validate(raw_classification).model_dump(mode="python")

    raw_consensus = value.get("consensus")
    if isinstance(raw_consensus, dict):
        normalized = {
            "classification": classification,
            "consensus": raw_consensus,
            "usage": value.get("usage", {}),
        }
        return normalized

    raw_likelihood = value.get("likelihood")
    likelihood = float(raw_likelihood) if isinstance(raw_likelihood, (int, float)) else None

    raw_votes = value.get("votes")
    choices: list[dict[str, Any]] = []
    if likelihood is not None or isinstance(raw_votes, list):
        choices.append({"classification": classification})
    if isinstance(raw_votes, list):
        for raw_vote in raw_votes:
            if not isinstance(raw_vote, dict):
                continue
            choices.append(
                {
                    "classification": ClassifyVote.model_validate(raw_vote).model_dump(mode="python"),
                }
            )

    return {
        "classification": classification,
        "consensus": {
            "choices": choices,
            "likelihood": likelihood,
        },
        "usage": value.get("usage", {}),
    }


class ClassifyResponse(BaseModel):
    classification: ClassifyResult = Field(..., description="The classification result with reasoning")
    consensus: ClassifyConsensus = Field(default_factory=ClassifyConsensus, description="Consensus metadata for multi-vote classification runs")
    usage: RetabUsage = Field(..., description="Usage information for the classification")

    @model_validator(mode="before")
    @classmethod
    def _coerce_legacy_classify_shape(cls, value: Any) -> Any:
        if not isinstance(value, dict):
            return value
        if "classification" in value and "usage" in value:
            return value
        return _build_native_classify_envelope(value)

    @property
    def result(self) -> ClassifyResult:
        return self.classification

    @property
    def likelihood(self) -> float | None:
        return self.consensus.likelihood

    @property
    def votes(self) -> list[ClassifyVote]:
        if not self.consensus.choices:
            return []

        choice_payloads = [choice.classification for choice in self.consensus.choices]
        if choice_payloads and (
            choice_payloads[0].category == self.classification.category
            and choice_payloads[0].reasoning == self.classification.reasoning
        ):
            return choice_payloads[1:]
        return choice_payloads


class ClassifyOutputSchema(BaseModel):
    """Schema for LLM structured output."""
    reasoning: str = Field(
        ..., 
        description="Step-by-step reasoning explaining why this document belongs to the chosen category"
    )
    classification: str = Field(
        ...,
        description="The category name that this document belongs to"
    )
