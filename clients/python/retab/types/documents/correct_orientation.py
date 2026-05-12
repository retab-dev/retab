from retab.types.base import RetabBaseModel

from ..mime import MIMEData


class DocumentTransformRequest(RetabBaseModel):
    document: MIMEData
    """The document to load."""


class DocumentTransformResponse(RetabBaseModel):
    document: MIMEData
    """The document to load."""
