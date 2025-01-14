from pydantic import BaseModel
from pathlib import Path
from typing import Any

class SchemaLoadingRequest(BaseModel):

    json_schema: dict[str, Any] | str | Path | None
    """The JSON schema to load."""
