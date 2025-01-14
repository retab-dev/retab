from typing import Literal


BaseModality = Literal["text", "image", "audio", "video"]
Modality = Literal[BaseModality, "native"]
TYPE_FAMILIES = Literal["excel", "word", "powerpoint", "pdf", "image", "text", "email", "audio"]
NativeModalities: dict[TYPE_FAMILIES, Modality] = {
    'excel': 'image',
    'word': 'image',
    'powerpoint': 'image',
    'pdf': 'image',
    'image': 'image',
    'text': 'text',
    'email': 'native',
    'audio': 'audio'
}

EXCEL_TYPES = Literal[".xls", ".xlsx", ".ods"]
WORD_TYPES = Literal[".doc", ".docx", ".odt"]
PPT_TYPES = Literal[".ppt", ".pptx", ".odp"]
PDF_TYPES = Literal[".pdf"]
IMAGE_TYPES = Literal[".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp"]
TEXT_TYPES = Literal[".txt", ".csv", ".tsv", ".md", ".log", ".html", ".htm", ".xml", ".json", ".yaml", ".yml", ".rtf", ".ini", ".conf", ".cfg", ".nfo", ".srt", ".sql", ".sh", ".bat", ".ps1", ".js", ".jsx", ".ts", ".tsx", ".py", ".java", ".c", ".cpp", ".cs", ".rb", ".php", ".swift", ".kt", ".go", ".rs", ".pl", ".r", ".m", ".scala"]
EMAIL_TYPES = Literal[".eml", ".msg"]
AUDIO_TYPES = Literal[".mp3", ".mp4", ".mpeg", ".mpga", ".m4a", ".wav", ".webm"]
SUPPORTED_TYPES = Literal[EXCEL_TYPES, WORD_TYPES, PPT_TYPES, PDF_TYPES, IMAGE_TYPES, TEXT_TYPES, EMAIL_TYPES, AUDIO_TYPES]
