from uiform import UiForm

uiclient = UiForm()
doc_msg = uiclient.documents.create_messages(
    document = "document_1.xlsx",
    modality = "native",
    text_operations = {
        "regex_instructions": [{
            "name": "VAT Number",
            "description": "All potential VAT numbers in the documents",
            "pattern": r"[Ff][Rr]\s*(\d\s*){11}"
        }],
    },
    image_operations = {
        "correct_image_orientation": True,
        "dpi": 72,
        "image_to_text": "ocr",
        "browser_canvas": "A4"
    }
)
