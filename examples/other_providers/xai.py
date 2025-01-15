from dotenv import load_dotenv
from uiform import UiForm, Schema
from pydantic import BaseModel, Field, ConfigDict
import os

assert load_dotenv()
print("UIFORM API KEY:",os.getenv("UIFORM_API_KEY"))

uiclient = UiForm()
doc_msg = uiclient.documents.create_messages(
    document = "freight/booking_confirmation.jpg"
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






from openai import OpenAI
client = OpenAI(
    api_key=os.getenv("XAI_API_KEY"),
    base_url="https://api.x.ai/v1"
)

#----

completion = client.chat.completions.create(
    model="grok-2-vision-1212",
    messages=doc_msg.openai_messages + [
        {
            "role": "user",
            "content": "Summarize the document"
        }
    ]
)

# --- OR 

completion = client.chat.completions.create(
    model="grok-2-vision-1212",
    messages=schema_obj.openai_messages + doc_msg.openai_messages,
    response_format={
        "type": "json_schema",
        "json_schema": {
            "name": schema_obj.schema_version,
            "schema": schema_obj.inference_json_schema,
            "strict": True
        }
    }
)

# --- OR 

completion = client.beta.chat.completions.parse(
    model="grok-2-vision-1212",
    messages=schema_obj.openai_messages + doc_msg.openai_messages,
    response_format=schema_obj.inference_pydantic_model
)


