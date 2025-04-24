from typing import Literal

from pydantic import BaseModel


class ImageSettings(BaseModel):
    correct_image_orientation: bool = True  # Whether to correct the image orientation
    dpi: int = 72  # The DPI of the image
    image_to_text: Literal["ocr", "llm_description"] = "ocr"  # Whether to convert the image to text
    browser_canvas: Literal['A3', 'A4', 'A5'] = "A4"  # The canvas of the browser (default = A4) - `A3`: 11.7in x 16.54in, `A4`: 8.27in x 11.7in, `A5`: 5.83in x 8.27in
