FROM python:3.11-slim

# 1) Install system dependencies and python packages
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/* \
 && pip install fastapi uvicorn pyngrok

# 2) Set working directory
WORKDIR /app

# 3) Write main.py (FastAPI app with ngrok integration)
RUN echo 'from fastapi import FastAPI, Request\n\
from pyngrok import ngrok\n\
import os\n\
import time\n\
\n\
app = FastAPI()\n\
\n\
@app.on_event("startup")\n\
async def startup_event():\n\
    # Configure ngrok\n\
    auth_token = os.getenv("NGROK_AUTH_TOKEN")\n\
    if not auth_token:\n\
        raise ValueError("NGROK_AUTH_TOKEN environment variable is required")\n\
    \n\
    ngrok.set_auth_token(auth_token)\n\
    \n\
    # Start ngrok tunnel\n\
    http_tunnel = ngrok.connect(8000)\n\
    public_url = http_tunnel.public_url\n\
    webhook_url = f"{public_url}/webhook"\n\
    \n\
    print("\\n🌍 Ngrok tunnel established!")\n\
    print(f"📬 Webhook URL: {webhook_url}\\n")\n\
    print("📬 Simple curl for testing: curl -X POST", webhook_url, "-H", "\"Content-Type: application/json\"", "-d", "'"'"'{\"message\": \"Hello, World!\"}'"'"'")\n\
\n\
@app.post("/webhook")\n\
async def webhook(request: Request):\n\
    data = await request.json()\n\
    print("📬 Webhook received:", data)\n\
    return {"status": "ok"}' > main.py

# 4) Final command
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]


### Usage:
# docker build -t fastapi-localtunnel . && docker run --rm -it -e NGROK_AUTH_TOKEN=[your-ngrok-auth-token] fastapi-localtunnel