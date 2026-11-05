"""Keep document bytes out of large processing requests."""

import base64
import json

from pydantic import BaseModel, Field

from retab.types.mime import MIMEData


MAX_INLINE_DOCUMENT_BYTES = 20 * 1024 * 1024
MAX_DOCUMENT_REQUEST_BYTES = 30 * 1024 * 1024


class InlineDocumentUpload(BaseModel):
    document_key: str | int | None
    filename: str
    content_type: str
    content: bytes = Field(repr=False)


def plan_document_uploads(data: object) -> list[InlineDocumentUpload]:
    if not isinstance(data, dict):
        return []
    candidates: list[tuple[str | int | None, object]] = [(None, data.get("document"))]
    documents = data.get("documents")
    if isinstance(documents, dict):
        candidates.extend((key, value) for key, value in documents.items() if isinstance(key, str))
    elif isinstance(documents, list):
        candidates.extend(enumerate(documents))
    inline: list[tuple[str | int | None, MIMEData]] = []
    for key, value in candidates:
        if not isinstance(value, dict):
            continue
        filename, url = value.get("filename"), value.get("url")
        if isinstance(filename, str) and isinstance(url, str) and url.startswith("data:"):
            inline.append((key, MIMEData(filename=filename, url=url)))
    if not inline:
        return []
    offload_all = len(json.dumps(data, ensure_ascii=False, separators=(",", ":")).encode()) > MAX_DOCUMENT_REQUEST_BYTES
    uploads: list[InlineDocumentUpload] = []
    for key, doc in inline:
        header, separator, encoded = doc.url[5:].partition(",")
        decoded_size = len(encoded) * 3 // 4 - (2 if encoded.endswith("==") else 1 if encoded.endswith("=") else 0)
        if not offload_all and decoded_size <= MAX_INLINE_DOCUMENT_BYTES:
            continue
        if not separator or not header.lower().endswith(";base64"):
            raise ValueError("Large inline documents must use base64 data URLs; upload the file first.")
        try:
            content = base64.b64decode(encoded, validate=True)
        except ValueError:
            raise ValueError("Large inline document contains invalid base64; upload the file first.") from None
        uploads.append(InlineDocumentUpload(document_key=key, filename=doc.filename, content_type=header[:-7] or "application/octet-stream", content=content))
    return uploads


def replace_uploaded_document(data: object, upload: InlineDocumentUpload, reference: MIMEData) -> dict[str, object]:
    if not isinstance(data, dict):
        raise ValueError("Document request must be an object")
    if not reference.url or reference.url.startswith("data:"):
        raise ValueError("Retab upload did not return a stored document reference")
    result = dict(data)
    if upload.document_key is None:
        result["document"] = {**data["document"], "filename": reference.filename, "url": reference.url}
    elif isinstance(upload.document_key, int):
        array = list(data["documents"])
        array[upload.document_key] = {**array[upload.document_key], "filename": reference.filename, "url": reference.url}
        result["documents"] = array
    else:
        documents = dict(data["documents"])
        documents[upload.document_key] = {**documents[upload.document_key], "filename": reference.filename, "url": reference.url}
        result["documents"] = documents
    return result


def check_document_request_size(data: object) -> None:
    if len(json.dumps(data, ensure_ascii=False, separators=(",", ":")).encode()) > MAX_DOCUMENT_REQUEST_BYTES:
        raise ValueError("Retab request exceeds 30 MiB after document upload; reduce JSON inputs.")
