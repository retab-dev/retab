from typing import Literal, List
from io import BytesIO, IOBase
from pathlib import Path
import PIL.Image

from ..._utils.mime import prepare_mime_document
from ..._resource import SyncAPIResource, AsyncAPIResource

from ...types.db.files import DBFile, FileData, FileTuple, FileLink
from ...types.mime import MIMEData
from ...types.standards import PreparedRequest

class FilesMixin:
    def prepare_download_link(self, file_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/db/files/{file_id}/download-link")
    
    def prepare_create(self,
                      document: Path | str | IOBase | MIMEData | PIL.Image.Image,
                      dataset_id: str | None = None,
                      dataset_name: str | None = None
                      ) -> PreparedRequest:
        if dataset_id is None and dataset_name is None:
            raise ValueError("Either dataset_id or dataset_name must be provided")
        if dataset_id is not None and dataset_name is not None:
            raise ValueError("Only one of dataset_id or dataset_name can be provided")

        mime_document = prepare_mime_document(document)

        content_binary = BytesIO(mime_document.content.encode('utf-8'))
        file_data: FileData = (mime_document.filename, content_binary, mime_document.mime_type)
        files: List[FileTuple] = [("file", file_data)]

        # Add dataset information as query parameters
        params = {}
        if dataset_id:
            params["dataset_id"] = dataset_id
        if dataset_name:
            params["dataset_name"] = dataset_name

        return PreparedRequest(method="POST", url="/v1/db/files", params=params)

    def prepare_get(self, file_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/db/files/{file_id}")

    def prepare_download(self, file_id: str) -> PreparedRequest:
        return PreparedRequest(method="GET", url=f"/v1/db/files/{file_id}/download")

    def prepare_delete(self, file_id: str) -> PreparedRequest:
        return PreparedRequest(method="DELETE", url=f"/v1/db/files/{file_id}")

    def prepare_list(self,
            dataset_id: str | None = None,
            mime_type: str | None = None,
            filename: str | None = None,
            file_id: str | None = None,
            after: str | None = None,
            before: str | None = None,
            limit: int = 10,
            order: Literal["asc", "desc"] | None = "desc"
                    ) -> PreparedRequest:
        params: dict[str, str | int] = {"limit": limit}
        if dataset_id:
            params["dataset_id"] = dataset_id
        if mime_type:
            params["mime_type"] = mime_type
        if filename:
            params["filename"] = filename
        if file_id:
            params["file_id"] = file_id
        if after:
            params["after"] = after
        if before:
            params["before"] = before
        if order:
            params["order"] = order
        return PreparedRequest(method="GET", url="/v1/db/files", params=params)


class Files(SyncAPIResource, FilesMixin):
    """Files API wrapper"""

    def download_link(self, file_id: str) -> FileLink:
        """
        Get a signed URL for accessing a stored object.

        Args:
            file_id: ID of the file in storage

        Returns:
            FileLink containing:
                - download_url: Signed URL for downloading the file
                - expires_in: Expiration time of the URL
                - filename: Name of the file
        """
        request = self.prepare_download_link(file_id)
        response = self._client._prepared_request(request)
        return FileLink.model_validate(response)
    
    def create(self,
                    document: Path | str | IOBase | MIMEData | PIL.Image.Image,
                    dataset_id: str | None = None,
                    dataset_name: str | None = None
    ) -> DBFile:
        """Upload the file to the server. 
        We force the user to specify a dataset membership when the user creates a file, to avoid orphan files.
        But if the dataset is then deleted, the file will still be there, and will have to be deleted separately if that's what the user wants.

        Args:
            document: The file to upload
            dataset_id: The ID of the dataset to add the file to
            dataset_name: The name of the dataset to add the file to
        

        Returns:
            File: The created file object
        """
        request = self.prepare_create(document, dataset_id, dataset_name)
        response = self._client._prepared_request(request)

        return DBFile(**response)
    
    def get(self, file_id: str) -> DBFile:
        """Get a file by ID.
        
        Args:
            file_id: The ID of the file to retrieve
            
        Returns:
            DBFile: The file object
        """
        request = self.prepare_get(file_id)
        response = self._client._prepared_request(request)
        return DBFile(**response)
    
    def download(self, file_id: str) -> MIMEData:
        """Download a file's content by ID.
        
        Args:
            file_id: The ID of the file to download
            
        Returns:
            MIMEData: The file content and metadata as a MIMEData object
            
        """
        # First get the file metadata
        file = self.get(file_id)
        
        # Then download the content
        request = self.prepare_download(file.id)
        response = self._client._prepared_request(request)
        
        # Create and return MIMEData object
        # Assuming the response contains the base64 encoded content directly
        return MIMEData.model_validate(response)
    
    def delete(self, file_id: str) -> None:
        """Delete a file by ID.
        
        Args:
            file_id: The ID of the file to delete
            
        """
        request = self.prepare_delete(file_id)
        self._client._prepared_request(request)

    def list(
        self,
        dataset_id: str | None = None,
        mime_type: str | None = None,
        filename: str | None = None,
        file_id: str | None = None,
        after: str | None = None,
        before: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc"
    ) -> List[DBFile]:
        """List files with pagination support.

        Args:
            dataset_id: The ID of the dataset to list files from.
            after: An object ID that defines your place in the list. When provided,
                returns objects after this ID. For example, if you receive 100 objects
                ending with "obj_123", you can pass after="obj_123" to fetch the next batch.
            before: An object ID that defines your place in the list. When provided,
                returns objects before this ID. Similar to 'after' but in reverse.
            limit: Upper limit on the number of objects to return, between 1 and 100.
                Defaults to 10 if not specified.
            order: Sort order by creation time. Use "asc" for oldest first,
                "desc" for newest first. Defaults to "desc" if not specified.

        Returns:
            List[DBFile]: A list of file objects matching the query parameters.
        """
        request = self.prepare_list(dataset_id, mime_type, filename, file_id, after, before, limit, order)
        response = self._client._prepared_request(request)
        return [DBFile(**item) for item in response["items"]]


class AsyncFiles(AsyncAPIResource, FilesMixin):
    """Async Files API wrapper"""

    async def download_link(self, file_id: str) -> FileLink:
        """
        Get a signed URL for accessing a stored object.

        Args:
            file_id: ID of the file in storage

        Returns:
            FileLink containing:
                - download_url: Signed URL for downloading the file
                - expires_in: Expiration time of the URL
                - filename: Name of the file
        """
        request = self.prepare_download_link(file_id)
        response = await self._client._prepared_request(request)
        return FileLink.model_validate(response)
    
    async def create(self,
                    document: Path | str | IOBase | MIMEData | PIL.Image.Image,
                    dataset_id: str | None = None,
                    dataset_name: str | None = None
    ) -> DBFile:
        """Upload the file to the server. 
        We force the user to specify a dataset membership when the user creates a file, to avoid orphan files.
        But if the dataset is then deleted, the file will still be there, and will have to be deleted separately if that's what the user wants.

        Args:
            document: The file to upload
            dataset_id: The ID of the dataset to add the file to
            dataset_name: The name of the dataset to add the file to
        

        Returns:
            File: The created file object
        """
        request = self.prepare_create(document, dataset_id, dataset_name)
        response = await self._client._prepared_request(request)

        return DBFile(**response)
    
    async def get(self, file_id: str) -> DBFile:
        """Get a file by ID.
        
        Args:
            file_id: The ID of the file to retrieve
            
        Returns:
            DBFile: The file object
        """
        request = self.prepare_get(file_id)
        response = await self._client._prepared_request(request)
        return DBFile(**response)
    
    async def download(self, file_id: str) -> MIMEData:
        """Download a file's content by ID.
        
        Args:
            file_id: The ID of the file to download
            
        Returns:
            MIMEData: The file content and metadata as a MIMEData object
            
        """
        # First get the file metadata
        file = await self.get(file_id)
        
        # Then download the content
        request = self.prepare_download(file.id)
        response = await self._client._prepared_request(request)
        
        # Create and return MIMEData object
        # Assuming the response contains the base64 encoded content directly
        return MIMEData.model_validate(response)
    
    async def delete(self, file_id: str) -> None:
        """Delete a file by ID.
        
        Args:
            file_id: The ID of the file to delete
            
        """
        request = self.prepare_delete(file_id)
        await self._client._prepared_request(request)

    async def list(
        self,
        dataset_id: str | None = None,
        mime_type: str | None = None,
        filename: str | None = None,
        file_id: str | None = None,
        after: str | None = None,
        before: str | None = None,
        limit: int = 10,
        order: Literal["asc", "desc"] | None = "desc"
    ) -> List[DBFile]:
        """List files with pagination support.

        Args:
            dataset_id: The ID of the dataset to list files from.
            after: An object ID that defines your place in the list. When provided,
                returns objects after this ID. For example, if you receive 100 objects
                ending with "obj_123", you can pass after="obj_123" to fetch the next batch.
            before: An object ID that defines your place in the list. When provided,
                returns objects before this ID. Similar to 'after' but in reverse.
            limit: Upper limit on the number of objects to return, between 1 and 100.
                Defaults to 10 if not specified.
            order: Sort order by creation time. Use "asc" for oldest first,
                "desc" for newest first. Defaults to "desc" if not specified.

        Returns:
            List[DBFile]: A list of file objects matching the query parameters.
        """
        request = self.prepare_list(dataset_id, mime_type, filename, file_id, after, before, limit, order)
        response = await self._client._prepared_request(request)
        return [DBFile(**item) for item in response["items"]]
