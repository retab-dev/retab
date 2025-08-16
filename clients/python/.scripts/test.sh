# Check for environment file at the specific required location
ENV_FILE_PATH="../../.env.local"
if [ -f "$ENV_FILE_PATH" ]; then
    echo "✅ Found .env.local file at open-source/sdk/.env.local"
    python3 -m pytest tests/ -v --tb=short --env-file="$ENV_FILE_PATH"
else
    echo "❌ ERROR: Environment file not found at open-source/sdk/.env.local"
    echo "   Please ensure the .env.local file exists at the correct location."
    echo "   Expected path: $(pwd)/$ENV_FILE_PATH"
    exit 1
fi