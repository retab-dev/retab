import hashlib
import mimetypes
from io import IOBase
from pathlib import Path
from typing import Any, List, Literal, Optional

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.files import CreateUploadResponse, File, FileLink, UploadFileResponse
from ...types.standards import PreparedRequest

FileUploadInput = Path | str | IOBase


class FilesMixin:
    def prepare_upload(
        self,
        filename: str,
        content_type: str,
        size_bytes: int,
        sha256: str | None = None,
    ) -> PreparedRequest:
        data: dict[str, Any] = {
            "filename": filename,
            "content_type": content_type,
            "size_bytes": size_bytes,
        }
        if sha256 is not None:
            data["sha256"] = sha256
        return PreparedRequest(
            method="POST",
            url="/files/upload",
            data=data,
        )

    def prepare_complete_upload(self, file_id: str, sha256: str | None = None) -> PreparedRequest:
        data: dict[str, Any] = {}
        if sha256 is not None:
            data["sha256"] = sha256
        return PreparedRequest(method="POST", url=f"/files/upload/{file_id}/complete", data=data)

    def prepare_list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: int = 10,
        order: Literal["asc", "desc"] = "desc",
        filename: Optional[str] = None,
        mime_type: Optional[str] = None,
        sort_by: str = "created_at",
    ) -> PreparedRequest:
        params: dict[str, Any] = {"limit": limit, "order": order, "sort_by": sort_by}
        if before is not None:
            params["before"] = before
        if after is not None:
            params["after"] = after
        if filename is not None:
            params["filename"] = filename
        if mime_type is not None:
            params["mime_type"] = mime_type
        return PreparedRequest(method="GET", url="/files", params=params)

    def prepare_get(self, file_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/files/{file_id}")

    def prepare_get_download_link(self, file_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/files/{file_id}/download-link")


def _is_remote_string(value: object) -> bool:
    return isinstance(value, str) and (value.startswith("http://") or value.startswith("https://") or value.startswith("data:"))


def _filename_for_file_obj(file_obj: IOBase) -> str:
    name = getattr(file_obj, "name", None)
    if isinstance(name, str) and name:
        return Path(name).name
    return "attachment"


def _content_type_for_filename(filename: str) -> str:
    return mimetypes.guess_type(filename)[0] or "application/octet-stream"


def _file_size_and_sha256(file_obj: IOBase) -> tuple[int, str]:
    if not file_obj.seekable():
        raise ValueError("files.upload requires a seekable file object so the upload size can be declared")
    current_position = file_obj.tell()
    digest = hashlib.sha256()
    size = 0
    file_obj.seek(0)
    while True:
        chunk = file_obj.read(1024 * 1024)
        if not chunk:
            break
        digest.update(chunk)
        size += len(chunk)
    file_obj.seek(current_position)
    return size, digest.hexdigest()


async def _async_file_chunks(file_obj: IOBase):
    import asyncio

    while True:
        chunk = await asyncio.to_thread(file_obj.read, 1024 * 1024)
        if not chunk:
            break
        yield chunk



class Files(SyncAPIResource, FilesMixin):

    def upload(self, mime_data: FileUploadInput) -> UploadFileResponse:
        if isinstance(mime_data, Path) or (isinstance(mime_data, str) and not _is_remote_string(mime_data)):
            path = Path(mime_data)
            with path.open("rb") as file_obj:
                response = self._upload_file_obj(
                    file_obj=file_obj,
                    filename=path.name,
                    content_type=_content_type_for_filename(path.name),
                )
            return UploadFileResponse(**response)
        if isinstance(mime_data, IOBase):
            filename = _filename_for_file_obj(mime_data)
            response = self._upload_file_obj(
                file_obj=mime_data,
                filename=filename,
                content_type=_content_type_for_filename(filename),
            )
            return UploadFileResponse(**response)
        raise ValueError("files.upload only accepts local file paths or seekable file objects")

    def _upload_file_obj(self, file_obj: IOBase, filename: str, content_type: str) -> dict[str, Any]:
        size_bytes, sha256 = _file_size_and_sha256(file_obj)
        session = CreateUploadResponse(**self._client._prepared_request(
            self.prepare_upload(
                filename=filename,
                content_type=content_type,
                size_bytes=size_bytes,
                sha256=sha256,
            )
        ))
        file_obj.seek(0)
        upload_response = self._client.client.put(
            session.upload_url,
            content=file_obj,
            headers=session.upload_headers,
        )
        upload_response.raise_for_status()
        return UploadFileResponse(**self._client._prepared_request(self.prepare_complete_upload(session.file_id, sha256=sha256))).model_dump()

    def list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: int = 10,
        order: Literal["asc", "desc"] = "desc",
        filename: Optional[str] = None,
        mime_type: Optional[str] = None,
        sort_by: str = "created_at",
    ) -> List[File]:
        request = self.prepare_list(before=before, after=after, limit=limit, order=order, filename=filename, mime_type=mime_type, sort_by=sort_by)
        response = self._client._prepared_request(request)
        return [File(**item) for item in response.get("data", [])]

    def get(self, file_id: str) -> File:
        request = self.prepare_get(file_id)
        response = self._client._prepared_request(request)
        return File(**response)

    def get_download_link(self, file_id: str) -> FileLink:
        request = self.prepare_get_download_link(file_id)
        response = self._client._prepared_request(request)
        return FileLink(**response)



class AsyncFiles(AsyncAPIResource, FilesMixin):

    async def upload(self, mime_data: FileUploadInput) -> UploadFileResponse:
        if isinstance(mime_data, Path) or (isinstance(mime_data, str) and not _is_remote_string(mime_data)):
            path = Path(mime_data)
            with path.open("rb") as file_obj:
                response = await self._upload_file_obj(
                    file_obj=file_obj,
                    filename=path.name,
                    content_type=_content_type_for_filename(path.name),
                )
            return UploadFileResponse(**response)
        if isinstance(mime_data, IOBase):
            filename = _filename_for_file_obj(mime_data)
            response = await self._upload_file_obj(
                file_obj=mime_data,
                filename=filename,
                content_type=_content_type_for_filename(filename),
            )
            return UploadFileResponse(**response)
        raise ValueError("files.upload only accepts local file paths or seekable file objects")

    async def _upload_file_obj(self, file_obj: IOBase, filename: str, content_type: str) -> dict[str, Any]:
        import asyncio

        size_bytes, sha256 = await asyncio.to_thread(_file_size_and_sha256, file_obj)
        session = CreateUploadResponse(**await self._client._prepared_request(
            self.prepare_upload(
                filename=filename,
                content_type=content_type,
                size_bytes=size_bytes,
                sha256=sha256,
            )
        ))
        file_obj.seek(0)
        upload_response = await self._client.client.put(
            session.upload_url,
            content=_async_file_chunks(file_obj),
            headers=session.upload_headers,
        )
        upload_response.raise_for_status()
        return UploadFileResponse(**await self._client._prepared_request(self.prepare_complete_upload(session.file_id, sha256=sha256))).model_dump()

    async def list(
        self,
        before: Optional[str] = None,
        after: Optional[str] = None,
        limit: int = 10,
        order: Literal["asc", "desc"] = "desc",
        filename: Optional[str] = None,
        mime_type: Optional[str] = None,
        sort_by: str = "created_at",
    ) -> List[File]:
        request = self.prepare_list(before=before, after=after, limit=limit, order=order, filename=filename, mime_type=mime_type, sort_by=sort_by)
        response = await self._client._prepared_request(request)
        return [File(**item) for item in response.get("data", [])]

    async def get(self, file_id: str) -> File:
        request = self.prepare_get(file_id)
        response = await self._client._prepared_request(request)
        return File(**response)

    async def get_download_link(self, file_id: str) -> FileLink:
        request = self.prepare_get_download_link(file_id)
        response = await self._client._prepared_request(request)
        return FileLink(**response)
