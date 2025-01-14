from typing import Any, Optional, Iterator, AsyncIterator
from types import TracebackType
import os
import httpx
import json
from pydantic_core import PydanticUndefined


from .resources import datasets, documents, files, finetuning, models, prompt_optimization, schemas

class BaseUiForm:
    """Base class for UiForm clients that handles authentication and configuration.

    Args:
        api_key (Optional[str]): UiForm API key. If not provided, will look for UIFORM_API_KEY env variable.
        base_url (Optional[str]): Base URL for API requests. Defaults to https://api.uiform.com
        timeout (float): Request timeout in seconds. Defaults to 240.0
        max_retries (int): Maximum number of retries for failed requests. Defaults to 3
        openai_api_key (Optional[str]): OpenAI API key. Will look for OPENAI_API_KEY env variable if not provided
        claude_api_key (Optional[str]): Claude API key. Will look for CLAUDE_API_KEY env variable if not provided
        xai_api_key (Optional[str]): XAI API key. Will look for XAI_API_KEY env variable if not provided
        gemini_api_key (Optional[str]): Gemini API key. Will look for GEMINI_API_KEY env variable if not provided
    """
    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: Optional[str] = None,
        timeout: float = 240.0,
        max_retries: int = 3,
        openai_api_key: Optional[str] = PydanticUndefined,  # type: ignore[assignment]
        claude_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
        xai_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
        gemini_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
    ) -> None:
        if api_key is None:
            api_key = os.environ.get("UIFORM_API_KEY")

        if api_key is None:
            raise ValueError(
                "The api_key client option must be set either by passing api_key to the client or by setting the UIFORM_API_KEY environment variable"
            )

        if base_url is None:    
            base_url = os.environ.get("UIFORM_API_BASE_URL", "https://api.uiform.com")

        self.api_key = api_key
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.max_retries = max_retries
        self.headers = {
            "Api-Key": self.api_key,
            "Content-Type": "application/json",
        }

        if openai_api_key is PydanticUndefined:
            openai_api_key = os.environ.get("OPENAI_API_KEY")

        if claude_api_key is PydanticUndefined:
            claude_api_key = os.environ.get("CLAUDE_API_KEY")
        
        if xai_api_key is PydanticUndefined:
            xai_api_key = os.environ.get("XAI_API_KEY")

        if gemini_api_key is PydanticUndefined:
            gemini_api_key = os.environ.get("GEMINI_API_KEY")

        if openai_api_key:
            self.headers["OpenAI-Api-Key"] = openai_api_key

        if claude_api_key:
            self.headers["Claude-Api-Key"] = claude_api_key

        if xai_api_key:
            self.headers["XAI-Api-Key"] = xai_api_key

        if gemini_api_key:
            self.headers["Gemini-Api-Key"] = gemini_api_key


class UiForm(BaseUiForm):
    """Synchronous client for interacting with the UiForm API.
    
    Provides access to all UiForm API resources through synchronous methods.
    """
    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: Optional[str] = None,
        timeout: float = 240.0,
        max_retries: int = 3,
        openai_api_key: Optional[str] = PydanticUndefined,  # type: ignore[assignment]
        claude_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
        xai_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
        gemini_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]

    ) -> None:
        super().__init__(
            api_key=api_key,
            base_url=base_url,
            timeout=timeout,
            max_retries=max_retries,
            openai_api_key=openai_api_key,
            claude_api_key=claude_api_key,
            xai_api_key=xai_api_key,
            gemini_api_key=gemini_api_key,
        )
        
        self.client = httpx.Client(timeout=self.timeout)

        self.files = files.Files(client=self)
        self.fine_tuning = finetuning.FineTuning(client=self)
        self.prompt_optimization = prompt_optimization.PromptOptimization(client=self)
        self.documents = documents.Documents(client=self)
        self.models = models.Models(client=self)
        self.datasets = datasets.Datasets(client=self)
        self.schemas = schemas.Schemas(client=self)
        
    def _request(
        self, method: str, endpoint: str, data: Optional[dict[str, Any]] = None
    ) -> Any:
        """Makes a synchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.)
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload

        Returns:
            Any: Parsed JSON response

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """
        url = f"{self.base_url}/{endpoint.lstrip('/')}"
        retries = 0
        while retries <= self.max_retries:
            try:
                response = self.client.request(
                    method, url, json=data, headers=self.headers
                )
                response.raise_for_status()
                return response.json()
            except httpx.HTTPStatusError as e:
                if response.status_code in {500, 502, 503, 504}:
                    retries += 1
                elif response.status_code == 422:
                    # Unprocessable Entity, likely a validation error
                    # Show the details in the exception message
                    raise RuntimeError(f"Validation error: {response.json()}")
                else:
                    raise e
            
            except httpx.RequestError as e:
                raise RuntimeError(f"Request failed: {e}")

        raise RuntimeError(f"Failed after {self.max_retries} retries")

    def _request_stream(
            self, method: str, endpoint: str, data: Optional[dict[str, Any]] = None
    ) -> Iterator[Any]:
        """Makes a streaming synchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.) 
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload

        Returns:
            Iterator[Any]: Generator yielding parsed JSON objects from the stream

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """
        url = f"{self.base_url}/{endpoint.lstrip('/')}"
        retries = 0
        # Returns a generator of JSON objects or raises an exception
        while retries < self.max_retries:
            with self.client.stream(method, url, json=data, headers=self.headers) as response_ctx_manager:
                try:
                    response_ctx_manager.raise_for_status()
                    # Iterate over the response stream and parse each line as JSON
                    for chunk in response_ctx_manager.iter_lines():
                        if not chunk: continue
                        try: yield json.loads(chunk)
                        except Exception: pass
                    # Once we finish the stream, we can break out of the loop
                    break
                except httpx.HTTPStatusError as e:
                    if response_ctx_manager.status_code in {500, 502, 503, 504}:
                        retries += 1
                    elif response_ctx_manager.status_code == 422:
                        # Unprocessable Entity, likely a validation error
                        # Show the details in the exception message
                        raise RuntimeError(f"Validation error: {response_ctx_manager.text}")
                    else:
                        raise e
        else:
            # If we did not break out of the loop, we reached the max retries
            raise RuntimeError(f"Failed after {self.max_retries} retries")

    def close(self) -> None:
        """Closes the HTTP client session."""
        self.client.close()

    def __enter__(self) -> "UiForm":
        """Context manager entry point.

        Returns:
            UiForm: The client instance
        """
        return self

    def __exit__(self, exc_type: type[BaseException] | None, exc_value: BaseException | None, traceback: TracebackType | None) -> None:
        """Context manager exit point that ensures the client is properly closed.

        Args:
            exc_type: The type of the exception that was raised, if any
            exc_value: The instance of the exception that was raised, if any
            traceback: The traceback of the exception that was raised, if any
        """
        self.close()



class AsyncUiForm(BaseUiForm):
    """Asynchronous client for interacting with the UiForm API.
    
    Provides access to all UiForm API resources through async methods.
    """
    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: Optional[str] = None,
        timeout: float = 240.0,
        max_retries: int = 3,
        openai_api_key: Optional[str] = PydanticUndefined,  # type: ignore[assignment]
        claude_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
        xai_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
        gemini_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
        ) -> None:
        super().__init__(
            api_key=api_key,
            base_url=base_url,
            timeout=timeout,
            max_retries=max_retries,
            openai_api_key=openai_api_key,
            claude_api_key=claude_api_key,
            xai_api_key=xai_api_key,
            gemini_api_key=gemini_api_key,
        )
        
        self.client = httpx.AsyncClient(timeout=self.timeout)

        self.files = files.AsyncFiles(client=self)
        self.fine_tuning = finetuning.AsyncFineTuning(client=self)
        self.prompt_optimization = prompt_optimization.AsyncPromptOptimization(client=self)
        self.documents = documents.AsyncDocuments(client=self)
        self.models = models.AsyncModels(client=self)
        self.datasets = datasets.AsyncDatasets(client=self)
        self.schemas = schemas.AsyncSchemas(client=self)

    async def _request(
        self, method: str, endpoint: str, data: Optional[dict[str, Any]] = None
    ) -> Any:
        """Makes an asynchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.)
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload

        Returns:
            Any: Parsed JSON response

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """
        url = f"{self.base_url}/{endpoint.lstrip('/')}"
        retries = 0
        while retries <= self.max_retries:
            try:
                response = await self.client.request(
                    method, url, json=data, headers=self.headers
                )
                response.raise_for_status()
                return response.json()
            except httpx.HTTPStatusError as e:
                if response.status_code in {500, 502, 503, 504}:
                    retries += 1
                elif response.status_code == 422:
                    raise RuntimeError(f"Validation error: {response.json()}")
                else:
                    raise e
            except httpx.RequestError as e:
                raise RuntimeError(f"Request failed: {e}")

        raise RuntimeError(f"Failed after {self.max_retries} retries")

    async def _request_stream(
        self, method: str, endpoint: str, data: Optional[dict[str, Any]] = None
    ) -> AsyncIterator[Any]:
        """Makes a streaming asynchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.)
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload

        Returns:
            AsyncIterator[Any]: Async generator yielding parsed JSON objects from the stream

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """
        url = f"{self.base_url}/{endpoint.lstrip('/')}"
        retries = 0

        while retries < self.max_retries:
            try:
                async with self.client.stream(method, url, json=data, headers=self.headers) as response_ctx_manager:
                    response_ctx_manager.raise_for_status()
                    async for line in response_ctx_manager.aiter_lines():
                        if not line:
                            continue
                        try:
                            yield json.loads(line)
                        except json.JSONDecodeError:
                            pass
                    return  # Exit after a successful streaming session
            except httpx.HTTPStatusError as e:
                if e.response.status_code in {500, 502, 503, 504}:
                    retries += 1
                elif e.response.status_code == 422:
                    raise RuntimeError(f"Validation error: {e.response.text}")
                else:
                    raise e
            except httpx.RequestError as e:
                raise RuntimeError(f"Request failed: {e}")

        raise RuntimeError(f"Failed after {self.max_retries} retries")

    async def close(self) -> None:
        """Closes the async HTTP client session."""
        await self.client.aclose()

    async def __aenter__(self) -> "AsyncUiForm":
        """Async context manager entry point.

        Returns:
            AsyncUiForm: The async client instance
        """
        return self

    async def __aexit__(self, exc_type: type, exc_value: BaseException, traceback: TracebackType) -> None:
        """Async context manager exit point that ensures the client is properly closed.

        Args:
            exc_type: The type of the exception that was raised, if any
            exc_value: The instance of the exception that was raised, if any
            traceback: The traceback of the exception that was raised, if any
        """
        await self.close()
