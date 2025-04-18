from uiform import UiForm

uiclient = UiForm()
doc_msg = uiclient.documents.create_messages(
    document = "document_1.xlsx",
    modality = "native",
    image_settings = {
        "correct_image_orientation": True,
        "dpi": 72,
        "image_to_text": "ocr",
        "browser_canvas": "A4"
    }
)
