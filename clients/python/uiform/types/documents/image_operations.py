from pydantic import BaseModel
from typing import TypedDict,Optional

    # 'A human-readable explanation of what is being extracted'

class ImageOperations(TypedDict, total=False):
    correct_orientation: bool
