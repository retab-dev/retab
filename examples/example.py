from uiform import UiForm
from openai import OpenAI

uiclient = UiForm()

doc_msg = uiclient.documents.create_messages(document = "freight/booking_confirmation.jpg", modality="text")

client = OpenAI()
completion = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=doc_msg.openai_messages + [
        {
            "role": "user",
            "content": "Summarize the document"
        }
    ]
)


print(completion.choices[0].message.content)
