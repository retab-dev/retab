from enum import Enum
from typing import Optional
from datetime import datetime
from pydantic import BaseModel
from ...types.ai_model import AIProvider


class ExternalAPIKey(BaseModel):
    """Request model for creating/updating API keys"""
    provider: AIProvider
    api_key: str

class ExternalAPIKeyResponse(BaseModel):
    """Response model for API key information"""
    provider: AIProvider
    is_configured: bool
    last_updated: Optional[datetime]
    api_key: Optional[str] = None
