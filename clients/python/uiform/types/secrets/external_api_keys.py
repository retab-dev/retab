from datetime import datetime
from enum import Enum
from typing import Optional

from pydantic import BaseModel

from ..ai_models import AIProvider


class ExternalAPIKeyRequest(BaseModel):
    """Request model for creating/updating API keys"""

    provider: AIProvider
    api_key: str


class ExternalAPIKey(BaseModel):
    """Response model for API key information"""

    provider: AIProvider
    is_configured: bool
    last_updated: Optional[datetime]
