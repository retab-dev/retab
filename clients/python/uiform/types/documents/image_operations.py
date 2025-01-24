from pydantic import BaseModel
from typing import TypedDict,Optional, Literal


class ImageOperations(TypedDict, total=False):
    correct_image_orientation: bool # Whether to correct the image orientation
    dpi : int # The DPI of the image
    image_to_text: Literal["ocr", "llm_description"] # Whether to convert the image to text
    browser_canvas: Literal['A3', 'A4', 'A5'] # The canvas of the browser (default = A4) - `A3`: 11.7in x 16.54in, `A4`: 8.27in x 11.7in, `A5`: 5.83in x 8.27in