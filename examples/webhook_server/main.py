from fastapi import FastAPI
from pyngrok import ngrok   # type: ignore
import os
import json
from uiform.types.automations.webhooks import WebhookRequest
from uiform.types.documents.extractions import UiParsedChatCompletion, UiParsedChoice
from openai.types.chat.parsed_chat_completion import ParsedChatCompletionMessage
from uiform.types.mime import MIMEData

app = FastAPI()

@app.on_event("startup")
async def startup_event():
    # Configure ngrok
    auth_token = os.getenv("NGROK_AUTH_TOKEN")
    if not auth_token:
        raise ValueError("NGROK_AUTH_TOKEN environment variable is required")
    
    ngrok.set_auth_token(auth_token)
    
    # Start ngrok tunnel
    http_tunnel = ngrok.connect("8000")
    public_url = http_tunnel.public_url
    webhook_url = f"{public_url}/webhook"

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
                    message=ParsedChatCompletionMessage(content="{\"message\" : \"Hello, World!\"}", role="assistant"),
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
    example_body_json = example_body.model_dump_json(exclude_unset=True, exclude_defaults=True)
    
    print("🌍 Ngrok tunnel established!")
    print(f"📬 Webhook URL: {webhook_url}")
    print("📬 Simple curl for testing:")
    print("-"*100)
    print("curl -X POST", webhook_url, "-H", "\"Content-Type: application/json\"", "-d", f"'{example_body_json}'")
    print("-"*100)


@app.post("/webhook")
async def webhook(request: WebhookRequest):
    invoice_object = json.loads(request.completion.choices[0].message.content or "{}") # The parsed object is the same Invoice object as the one you defined in the Pydantic model
    print("📬 Webhook received:", invoice_object)
    return {"status": "success", "data": invoice_object}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)

# To run the FastAPI app locally, use the command:
# uvicorn your_module_name:app --reload