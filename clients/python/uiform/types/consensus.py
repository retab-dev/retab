from pydantic import BaseModel, Field


class ReconciliationResponse(BaseModel):
    consensus_dict: dict = Field(
        description="The consensus dictionary containing the reconciled values from the input dictionaries."
    )
    likelihoods: dict = Field(
        description="A dictionary containing the likelihood/confidence scores for each field in the consensus dictionary."
    )
