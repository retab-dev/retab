from typing import Literal

BaseModality = Literal["text", "image"]  # "video" , "audio"
Modality = Literal[BaseModality, "native"]