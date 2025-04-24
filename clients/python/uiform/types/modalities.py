from typing import Literal

BaseModality = Literal["text", "image"]  # "video" , "audio"
Modality = Literal[BaseModality, "native", "image+text"]
TYPE_FAMILIES = Literal["excel", "word", "powerpoint", "pdf", "image", "text", "email", "audio", "html", "web"]
NativeModalities: dict[TYPE_FAMILIES, Modality] = {
    'excel': 'image',
    'word': 'image',
    'html': 'text',
    'powerpoint': 'image',
    'pdf': 'image',
    'image': 'image',
    'web': 'image',
    'text': 'text',
    'email': 'native',
    'audio': 'text',
}

EXCEL_TYPES = Literal[".xls", ".xlsx", ".ods"]
WORD_TYPES = Literal[".doc", ".docx", ".odt"]
PPT_TYPES = Literal[".ppt", ".pptx", ".odp"]
PDF_TYPES = Literal[".pdf"]
IMAGE_TYPES = Literal[".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp"]
TEXT_TYPES = Literal[
    ".txt",
    ".csv",
    ".tsv",
    ".md",
    ".log",
    ".xml",
    ".json",
    ".yaml",
    ".yml",
    ".rtf",
    ".ini",
    ".conf",
    ".cfg",
    ".nfo",
    ".srt",
    ".sql",
    ".sh",
    ".bat",
    ".ps1",
    ".js",
    ".jsx",
    ".ts",
    ".tsx",
    ".py",
    ".java",
    ".c",
    ".cpp",
    ".cs",
    ".rb",
    ".php",
    ".swift",
    ".kt",
    ".go",
    ".rs",
    ".pl",
    ".r",
    ".m",
    ".scala",
]
HTML_TYPES = Literal[".html", ".htm"]
WEB_TYPES = Literal[".mhtml"]
EMAIL_TYPES = Literal[".eml", ".msg"]
AUDIO_TYPES = Literal[".mp3", ".mp4", ".mpeg", ".mpga", ".m4a", ".wav", ".webm"]
SUPPORTED_TYPES = Literal[EXCEL_TYPES, WORD_TYPES, PPT_TYPES, PDF_TYPES, IMAGE_TYPES, TEXT_TYPES, HTML_TYPES, WEB_TYPES, EMAIL_TYPES, AUDIO_TYPES]
