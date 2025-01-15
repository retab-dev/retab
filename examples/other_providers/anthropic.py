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



from anthropic import Anthropic
client = Anthropic(api_key=os.getenv("CLAUDE_API_KEY"))

#----

completion = client.messages.create(
    model="claude-3-5-sonnet-20241022",
    max_tokens=1000,
    temperature=0,
    system=doc_msg.anthropic_system_prompt,
    messages=doc_msg.anthropic_messages + [
        {
            "role": "user",
            "content": "Summarize this Document"
        }
    ]
)

# --- OR 

# Using Anthropic's JSON mode
completion = client.messages.create(
    model="claude-3-5-sonnet-20241022",
    max_tokens=1000,
    temperature=0,
    system=schema_obj.anthropic_system_prompt + doc_msg.anthropic_system_prompt,
    messages=schema_obj.anthropic_messages + doc_msg.anthropic_messages
)

from anthropic.types.text_block import TextBlock
assert isinstance(completion.content[0], TextBlock)
print(completion.content[0].text)
