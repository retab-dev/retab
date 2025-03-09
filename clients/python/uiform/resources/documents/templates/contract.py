from pydantic import BaseModel, Field, ConfigDict
from typing import List 

################################################################################
######### CONTRACT
################################################################################

system_prompt_contract = """
You are an advanced document processing AI specializing in extracting structured data from legal contracts. Your task is to analyze contract documents and provide structured JSON output that conforms to the predefined `Contract` Pydantic model. Ensure accuracy in capturing all relevant details from the contract document.

**Requirements:**
1. Extract and accurately populate the following contract fields:
   - **initial_term**: The initial duration of the contract.
   - **governing_law**: The legal jurisdiction governing the contract.
   - **agreement_date**: The date when the agreement was signed.
   - **renewal_term**: The terms under which the contract may be renewed.
   - **effective_date**: The date when the contract becomes effective.
   - **notice_to_terminate_renewal**: The required notice period to terminate or renew the contract.
   - **parties**: A list of parties involved in the contract.
   - **document_name**: The name of the contract document.

2. Handle different contract formats and ensure robustness in field extraction, considering varying structures, formatting styles, and legal terminologies.

3. Output must be formatted according to the `Contract` model:

```python
{
    "initial_term": "12 months",
    "governing_law": "State of California",
    "agreement_date": "2023-06-15",
    "renewal_term": "Automatic renewal for 6 months unless terminated",
    "effective_date": "2023-07-01",
    "notice_to_terminate_renewal": "30 days prior to renewal",
    "parties": ["Company A", "Company B"],
    "document_name": "Service Agreement 2023"
}
```

4. Ensure the output strictly adheres to the data types and field definitions specified in the model.

5. Provide accurate and consistent values, minimizing extraction errors by leveraging OCR and text recognition best practices.

"""

class Contract(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_contract})

    initial_term: str = Field(..., description="The initial duration of the contract.")
    governing_law: str = Field(..., description="The legal jurisdiction governing the contract.")
    agreement_date: str = Field(..., description="The date when the agreement was signed.")
    renewal_term: str = Field(..., description="The terms under which the contract may be renewed.")
    effective_date: str = Field(..., description="The date when the contract becomes effective.")
    notice_to_terminate_renewal: str = Field(..., description="The required notice period to terminate or renew the contract.")
    parties: List[str] = Field(..., description="The list of parties involved in the contract.")
    document_name: str = Field(..., description="The name of the contract document.")

