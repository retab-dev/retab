from pydantic import BaseModel, Field, ConfigDict
from typing import Optional

################################################################################
######### IDENTITY PROOFING
################################################################################

system_prompt_identity_proofing = """
You are an expert document analysis AI specializing in identity proofing. Your task is to analyze identity documents and extract relevant fraud detection signals and evidence. 

**Your output should be structured according to the `IdentityProofingAnalysis` Pydantic model, with the following requirements:**

1. **Fraud Signals:** Identify potential fraud-related aspects of the document, including:
    - Whether the document is an identity document.
    - Suspicious words that may indicate fraud.
    - Any signs of image manipulation.
    - Whether the document appears duplicated online.
    - Detection of photocopies.

2. **Evidence:** Provide supporting details, such as:
    - A list of suspicious words found in the document.
    - Any inconclusive suspicious words requiring further review.
    - A URL to a thumbnail image of the document.
    - The hostname where the document was found online.

**Input Example:**

An image or scanned document of an identity card or passport.

**Expected Output:**

```python
IdentityProofing(
    fraud_signals=FraudSignals(
        is_identity_document=True,
        suspicious_words=["fake", "altered"],
        image_manipulation=True,
        online_duplicate=False,
        photocopy_detection=True
    ),
    evidence=Evidence(
        suspicious_words=["fake"],
        inconclusive_suspicious_word="altered",
        hostname="example.com"
    )
)
```

Ensure the output is structured, accurate, and aligns with the defined model while maintaining high reliability in fraud detection.

"""

class FraudSignals(BaseModel):
    is_identity_document: list[bool] = Field(default_factory=list, description="List of indicators if the document is identified as an identity document")
    suspicious_words: list[str] = Field(default_factory=list, description="List of words flagged as suspicious in the document")
    image_manipulation: list[bool] = Field(default_factory=list, description="List of flags indicating if the document image shows signs of manipulation")
    online_duplicate: list[bool] = Field(default_factory=list, description="List of flags indicating if the document appears duplicated online")
    photocopy_detection: list[bool] = Field(default_factory=list, description="List of flags indicating if the document is detected as a photocopy")

class Evidence(BaseModel):
    suspicious_words: list[str] = Field(default_factory=list, description="List of suspicious words found in the document")
    inconclusive_suspicious_word: list[Optional[str]] = Field(default_factory=list, description="List of words flagged as suspicious but inconclusive")
    hostname: list[Optional[str]] = Field(default_factory=list, description="List of hostnames where the document was found online")

class IdentityProofing(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_identity_proofing})
    
    fraud_signals: FraudSignals = Field(..., description="Fraud detection signals related to the document")
    evidence: Evidence = Field(..., description="Supporting evidence for the detected fraud signals")
