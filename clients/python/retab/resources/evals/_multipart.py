import base64
import json
from io import IOBase
from pathlib import Path
from typing import Any, Dict, Sequence

import PIL.Image
from pydantic import HttpUrl

from ...types.standards import PreparedRequest
from ...utils.mime import MIMEData, prepare_mime_document

DocumentInput = Path | str | bytes | IOBase | MIMEData | PIL.Image.Image | HttpUrl


def build_multipart_files(
    *,
    document: DocumentInput | None = None,
    documents: Sequence[DocumentInput] | None = None,
) -> dict[str, Any] | list[tuple[str, tuple[str, bytes, str]]]:
    if not document and not documents:
        raise ValueError("Either 'document' or 'documents' must be provided")
    if document and documents:
        raise ValueError("Provide either 'document' (single) or 'documents' (multiple), not both")

    if document:
        mime_document = prepare_mime_document(document)
        return {
            "document": (mime_document.filename, base64.b64decode(mime_document.content), mime_document.mime_type)
        }

    files: list[tuple[str, tuple[str, bytes, str]]] = []
    assert documents is not None
    for item in documents:
        mime_document = prepare_mime_document(item)
        files.append(
            (
                "documents",
                (mime_document.filename, base64.b64decode(mime_document.content), mime_document.mime_type),
            )
        )
    return files


def build_form_data(
    *,
    model: str | None = None,
    image_resolution_dpi: int | None = None,
    n_consensus: int | None = None,
    metadata: Dict[str, str] | None = None,
    extraction_id: str | None = None,
    **extra_form: Any,
) -> dict[str, Any]:
    form_data = {
        "model": model,
        "image_resolution_dpi": image_resolution_dpi,
        "n_consensus": n_consensus,
        "metadata": json.dumps(metadata) if metadata else None,
        "extraction_id": extraction_id,
    }
    if extra_form:
        form_data.update(extra_form)
    return {key: value for key, value in form_data.items() if value is not None}


def build_process_request(
    *,
    base: str,
    eval_id: str,
    iteration_id: str | None = None,
    document: DocumentInput | None = None,
    documents: Sequence[DocumentInput] | None = None,
    model: str | None = None,
    image_resolution_dpi: int | None = None,
    n_consensus: int | None = None,
    metadata: Dict[str, str] | None = None,
    extraction_id: str | None = None,
    **extra_form: Any,
) -> PreparedRequest:
    form_data = build_form_data(
        model=model,
        image_resolution_dpi=image_resolution_dpi,
        n_consensus=n_consensus,
        metadata=metadata,
        extraction_id=extraction_id,
        **extra_form,
    )
    files = build_multipart_files(document=document, documents=documents)

    url = f"{base}/extract/{eval_id}" if iteration_id is None else f"{base}/extract/{eval_id}/{iteration_id}"
    return PreparedRequest(method="POST", url=url, form_data=form_data, files=files)
