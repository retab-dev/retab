from pydantic import BaseModel

from ..mime import MIMEData


class DocumentTransformRequest(BaseModel):
    document: MIMEData
    """The document to load."""


class DocumentTransformResponse(BaseModel):
    document: MIMEData
    """The document to load."""
