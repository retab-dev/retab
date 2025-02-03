from .._resource import SyncAPIResource, AsyncAPIResource
from ..types.standards import PreparedRequest

class ModelsMixin:
    def prepare_list(self) -> PreparedRequest:
        return PreparedRequest(
            method="GET",
            url="/v1/models"
        )

class Models(SyncAPIResource, ModelsMixin): 
    """Models API wrapper"""

    def list(self) -> list[str]:
        """
        List all available models.
        
        Returns:
            list[str]: List of available models
        Raises:
            HTTPException if the request fails
        """

        request = self.prepare_list()
        return self._client._prepared_request(request)

    
class AsyncModels(AsyncAPIResource, ModelsMixin): 
    """Models Asyncronous API wrapper"""
    
    async def list(self) -> list[str]:
        """
        List all available models.
        
        Returns:
            list[str]: List of available models
        Raises:
            HTTPException if the request fails
        """

        request = self.prepare_list()
        return await self._client._prepared_request(request)