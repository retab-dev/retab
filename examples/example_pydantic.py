from uiform import UiForm, Schema
from openai import OpenAI
from pydantic import BaseModel, Field, ConfigDict

uiclient = UiForm()
doc_msg = uiclient.documents.create_messages(
    document = "document_1.xlsx"
)

class CalendarEvent(BaseModel):
    model_config = ConfigDict(json_schema_extra = {"X-SystemPrompt": "You are a useful assistant."})

    name: str = Field(...,
        description="The name of the calendar event.",
        json_schema_extra={"X-FieldPrompt": "Provide a descriptive and concise name for the event."}
    )
    date: str = Field(...,
        description="The date of the calendar event in ISO 8601 format.",
        json_schema_extra={
            'X-ReasoningPrompt': 'The user can mention it in any format, like **next week** or **tomorrow**. Infer the right date format from the user input.',
        }
    )

schema_obj =Schema(
    pydantic_model = CalendarEvent
)

# Now you can use your favorite model to analyze your document
client = OpenAI()
completion = client.beta.chat.completions.parse(
    model="gpt-4o",
    messages=schema_obj.openai_messages + doc_msg.openai_messages,
    response_format=schema_obj.response_format_pydantic
)

# Validate the response against the original schema if you want to remove the reasoning fields
assert completion.choices[0].message.content is not None
extraction = schema_obj.pydantic_model.model_validate_json(
    completion.choices[0].message.content 
)

print("Extraction:",extraction)
