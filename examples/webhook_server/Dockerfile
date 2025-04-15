FROM python:3.12-slim

# 1) Install system dependencies and python packages
RUN apt-get update && apt-get install -y curl git && rm -rf /var/lib/apt/lists/*

# 2) Copy requirements.txt
COPY requirements.txt .

# 3) Install python packages
RUN pip install -r requirements.txt

# 4) Set working directory
WORKDIR /app

# 5) Copy main.py
COPY main.py main.py

# 4) Final command
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]


### Usage:
# docker build -t fastapi-ngrok . && docker run --rm -it -e NGROK_AUTH_TOKEN=[your-ngrok-auth-token] fastapi-ngrok