import google.generativeai as genai
from uiform import UiForm

uiclient = UiForm()
doc_msg = uiclient.documents.create_messages(
    document = "freight/booking_confirmation.jpg"
)

genai.configure()
model = genai.GenerativeModel("gemini-1.5-flash")
response = model.generate_content(
    doc_msg.gemini_messages + ["Summarize the document"]
)
