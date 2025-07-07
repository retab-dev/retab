FROM node:18-slim

# 1) Install system dependencies
RUN apt-get update && apt-get install -y curl git && rm -rf /var/lib/apt/lists/*

# 2) Set working directory
WORKDIR /app

# 3) Copy package files
COPY package.json package-lock.json* ./

# 4) Install dependencies
RUN npm install

# 5) Copy source files
COPY main.js ./

# 6) Expose port
EXPOSE 8000

# 7) Final command
CMD ["npm", "start"]

### Usage:
# docker build -t express-ngrok . && docker run --rm -it -e NGROK_AUTH_TOKEN=[your-ngrok-auth-token] -p 8000:8000 express-ngrok