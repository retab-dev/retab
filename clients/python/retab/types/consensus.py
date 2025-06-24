from typing import Any, Literal, Optional
from pydantic import BaseModel, Field


class ReconciliationRequest(BaseModel):
    list_dicts: list[dict] = Field(description="List of dictionaries that will be reconciled into a single consensus dictionary.")
    reference_schema: Optional[dict[str, Any]] = Field(
        default=None,
        description="Optional schema defining the structure and types of the dictionary fields to validate the list of dictionaries against. Raise an error if one of the dictionaries does not match the schema.",
    )
    mode: Literal["direct", "aligned"] = Field(
        default="direct",
        description="The mode to use for the consensus. If 'direct', the consensus is computed directly from the list of dictionaries. If 'aligned', the consensus is computed from the aligned dictionaries.",
    )


class ReconciliationResponse(BaseModel):
    consensus_dict: dict = Field(description="The consensus dictionary containing the reconciled values from the input dictionaries.")
    likelihoods: dict = Field(description="A dictionary containing the likelihood/confidence scores for each field in the consensus dictionary.")
