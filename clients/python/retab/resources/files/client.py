from typing import Any, List, Literal, Optional

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.files import File, FileLink, UploadFileResponse
from ...types.mime import MIMEData
from ...types.standards import PreparedRequest


class FilesMixin:
    def prepare_upload(self, mime_data: MIMEData) -> PreparedRequest:
        return PreparedRequest(
            method="POST",
            url="/v1/files/upload",
            data={"mimeData": mime_data.model_dump(mode="json")},
        )

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
        return PreparedRequest(method="GET", url="/v1/files", params=params)

    def prepare_get(self, file_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/files/{file_id}")

    def prepare_get_download_link(self, file_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/files/{file_id}/download-link")



class Files(SyncAPIResource, FilesMixin):

    def upload(self, mime_data: MIMEData) -> UploadFileResponse:
        request = self.prepare_upload(mime_data)
        response = self._client._prepared_request(request)
        return UploadFileResponse(**response)

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

    async def upload(self, mime_data: MIMEData) -> UploadFileResponse:
        request = self.prepare_upload(mime_data)
        response = await self._client._prepared_request(request)
        return UploadFileResponse(**response)

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

