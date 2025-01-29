from enum import Enum
from typing import Optional
from datetime import datetime
from pydantic import BaseModel

class ApiProvider(str, Enum):
    """Supported external API providers"""
    OPENAI = "openai"
    GEMINI = "gemini"
    ANTHROPIC = "anthropic"
    X_AI = "xai"

class ExternalAPIKey(BaseModel):
    """Request model for creating/updating API keys"""
    provider: ApiProvider
    api_key: str

class ExternalAPIKeyResponse(BaseModel):
    """Response model for API key information"""
    provider: ApiProvider
    is_configured: bool
    last_updated: Optional[datetime]
    api_key: Optional[str] = None
