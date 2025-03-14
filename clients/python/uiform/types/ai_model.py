from typing import Literal

AIProvider = Literal["OpenAI"]#, "Anthropic", "xAI", "Gemini"]
OpenAICompatibleProvider = Literal["OpenAI"]#, "xAI", "Gemini"]
GeminiModel = Literal[ "gemini-2.0-flash-exp",
                      "gemini-1.5-flash-8b", "gemini-1.5-flash","gemini-1.5-pro"]
AnthropicModel = Literal["claude-3-5-sonnet-latest","claude-3-5-sonnet-20241022",
                         "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"]
OpenAIModel = Literal["gpt-4o", "gpt-4o-mini","chatgpt-4o-latest",
                      "gpt-4o-2024-11-20", "gpt-4o-2024-08-06", "gpt-4o-2024-05-13",
                      "gpt-4o-mini-2024-07-18",
                      "o3-mini", "o3-mini-2025-01-31",
                      "o1", "o1-2024-12-17", "o1-mini", "o1-mini-2024-09-12",
                      "gpt-4.5-preview", "gpt-4.5-preview-2025-02-27"]
xAI_Model = Literal["grok-2-vision-1212", "grok-2-1212"]
LLMModel = Literal[OpenAIModel, 'human']# [AnthropicModel, OpenAIModel, xAI_Model, GeminiModel]


from pydantic import BaseModel, Field
import datetime

from uiform.types.jobs.base import AnnotationProps
class FinetunedModel(BaseModel):
    object: Literal["finetuned_model"] = "finetuned_model"
    organization_id: str
    model: str
    schema_id: str
    schema_data_id: str 
    finetuning_props : AnnotationProps
    created_at: datetime.datetime = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc))



# Example finetuned model instances
from uiform.types.modalities import Modality
from uiform.types.image_settings import ImageSettings

# Example 1: Document extraction model
document_finetuned_model = FinetunedModel(
    model="ft-document-extraction-v1",
    organization_id="org_123456",
    schema_id="schema_invoice_extraction",
    schema_data_id="schema_data_invoice_v1",
    finetuning_props=AnnotationProps(
        model="gpt-4o",
        temperature=0.1,
        modality="native",
        image_settings=ImageSettings()
    )
)

# Example 2: Image classification model
image_finetuned_model = FinetunedModel(
    model="ft-image-classification-v2",
    organization_id="org_789012",
    schema_id="schema_product_categorization",
    schema_data_id="schema_data_products_v3",
    finetuning_props=AnnotationProps(
        model="gpt-4o-mini",
        temperature=0.0,
        modality="image",
        image_settings=ImageSettings()
    )
)

# Example 3: Text analysis model
text_finetuned_model = FinetunedModel(
    model="ft-text-analysis-v1",
    organization_id="org_345678",
    schema_id="schema_sentiment_analysis",
    schema_data_id="schema_data_sentiment_v2",
    finetuning_props=AnnotationProps(
        model="o1-mini-2024-09-12",
        temperature=0.2,
        modality="text",
        image_settings=ImageSettings()
    )
)

# We should tell people to change their model for a finetuned model only if: 
# - The schema_id match
# - The finetuning_props modalities and image_settings are the same. 




