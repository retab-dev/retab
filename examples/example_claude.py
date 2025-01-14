from anthropic import Anthropic
from uiform import UiForm

uiclient = UiForm()

doc_msg = uiclient.documents.create_messages(document = "freight/booking_confirmation.jpg")

client = Anthropic()
message = client.messages.create(
    model="claude-3-5-sonnet-20241022",
    max_tokens=1000,
    temperature=0,
    messages=doc_msg.anthropic_messages + [
        {
            "role": "user",
            "content": "Summarize this Document"
        }
    ]
)
print(message.model_dump())