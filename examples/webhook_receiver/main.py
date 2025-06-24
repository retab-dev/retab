# ---------------------------------------------
## FastAPI app to receive Retab webhook events (with ngrok tunnel)
# ---------------------------------------------

import json
import os

from dotenv import load_dotenv
from fastapi import FastAPI
from openai.types.chat.parsed_chat_completion import ParsedChatCompletionMessage
from pyngrok import ngrok

from retab.types.automations.webhooks import WebhookRequest
from retab.types.documents.extractions import UiParsedChatCompletion, UiParsedChoice
from retab.types.mime import MIMEData

# Load environment variables
load_dotenv()

app = FastAPI()


@app.on_event("startup")
async def startup_event():
    auth_token = os.getenv("NGROK_AUTH_TOKEN")
    if not auth_token:
        raise ValueError("NGROK_AUTH_TOKEN environment variable is required")

    ngrok.set_auth_token(auth_token)
    http_tunnel = ngrok.connect("8000")
    public_url = http_tunnel.public_url
    webhook_url = f"{public_url}/webhook"

    # Print example curl
    example_body = WebhookRequest(
        completion=UiParsedChatCompletion(
            id="id",
            created=0,
            model="gpt-4.1-nano",
            object="chat.completion",
            likelihoods={},
            choices=[
                UiParsedChoice(
                    index=0,
                    message=ParsedChatCompletionMessage(content="{\"message\": \"Hello, World!\"}", role="assistant"),
                    finish_reason=None,
                    logprobs=None,
                )
            ],
        ),
        file_payload=MIMEData(
            filename="example.pdf",
            url="data:application/pdf;base64,the_content_of_the_pdf_file",
        ),
    )

    print("\n🌍 Ngrok tunnel established!")
    print(f"📬 Webhook URL: {webhook_url}")
    print("\n📬 Test with curl:")
    print("-" * 80)
    print(f"curl -X POST {webhook_url} -H \"Content-Type: application/json\" -d '{example_body.model_dump_json()}'")
    print("-" * 80)


@app.post("/webhook")
async def webhook(request: WebhookRequest):
    parsed_data = json.loads(request.completion.choices[0].message.content or "{}")
    print("\n✅ Webhook received:")
    print(json.dumps(parsed_data, indent=2))
    return {"status": "success", "data": parsed_data}


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
