from .._resource import SyncAPIResource, AsyncAPIResource


class Models(SyncAPIResource): 
    """Models API wrapper"""

    def list(self) -> list[str]:
        """
        List all available models.
        
        Returns:
            list[str]: List of available models
        Raises:
            HTTPException if the request fails
        """

        return self._client._request("GET", "/api/v1/models")["content"]

    
class AsyncModels(AsyncAPIResource): 
    """Models Asyncronous API wrapper"""
    
    async def list(self) -> list[str]:
        """
        List all available models.
        
        Returns:
            list[str]: List of available models
        Raises:
            HTTPException if the request fails
        """

        return (await self._client._request("GET", "/api/v1/models"))["content"]
