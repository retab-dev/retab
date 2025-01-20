from typing import Any, Optional, Iterator, AsyncIterator, BinaryIO, List
from types import TracebackType
import os
import httpx
import json
import backoff
import backoff.types
from pydantic_core import PydanticUndefined

from .types.files_datasets import FileTuple
from .resources import datasets, documents, files, finetuning, models, prompt_optimization, schemas, files_datasets

class MaxRetriesExceeded(Exception): pass


def raise_max_tries_exceeded(details: backoff.types.Details) -> None:
    exception = details.get("exception")
    tries = details["tries"]
    if isinstance(exception, BaseException):
        raise Exception(f"Max tries exceeded after {tries} tries.") from exception
    else:
        raise Exception(f"Max tries exceeded after {tries} tries.")

class BaseUiForm:
    """Base class for UiForm clients that handles authentication and configuration.

    This class provides core functionality for API authentication, configuration, and common HTTP operations
    used by both synchronous and asynchronous clients.

    Args:
        api_key (str, optional): UiForm API key. If not provided, will look for UIFORM_API_KEY env variable.
        base_url (str, optional): Base URL for API requests. Defaults to https://api.uiform.com
        timeout (float): Request timeout in seconds. Defaults to 240.0
        max_retries (int): Maximum number of retries for failed requests. Defaults to 3
        openai_api_key (str, optional): OpenAI API key. Will look for OPENAI_API_KEY env variable if not provided
        claude_api_key (str, optional): Claude API key. Will look for CLAUDE_API_KEY env variable if not provided
        xai_api_key (str, optional): XAI API key. Will look for XAI_API_KEY env variable if not provided
        gemini_api_key (str, optional): Gemini API key. Will look for GEMINI_API_KEY env variable if not provided

    Raises:
        ValueError: If no API key is provided through arguments or environment variables
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
                "No API key provided. You can create an API key at https://uiform.com\n"
                "Then either pass it to the client (api_key='your-key') or set the UIFORM_API_KEY environment variable"
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


    def _prepare_url(self, endpoint: str) -> str:
        return f"{self.base_url}/{endpoint.lstrip('/')}"
    
    def _validate_response(self, response_object: httpx.Response) -> None:
        if response_object.status_code in {500, 502, 503, 504}:
            response_object.raise_for_status()
        elif response_object.status_code == 422:
            raise RuntimeError(f"Validation error: {response_object.json()}")
        elif not response_object.is_success:
            raise RuntimeError(f"Request failed: {response_object.json()}")
        
    def _get_headers(self, idempotency_key: str | None = None) -> dict[str, Any]:
        headers = self.headers.copy()
        if idempotency_key:
            headers["Idempotency-Key"] = idempotency_key
        return headers

class UiForm(BaseUiForm):
    """Synchronous client for interacting with the UiForm API.
    
    This client provides synchronous access to all UiForm API resources including files, fine-tuning,
    prompt optimization, documents, models, datasets, and schemas.

    Args:
        api_key (str, optional): UiForm API key. If not provided, will look for UIFORM_API_KEY env variable.
        base_url (str, optional): Base URL for API requests. Defaults to https://api.uiform.com
        timeout (float): Request timeout in seconds. Defaults to 240.0
        max_retries (int): Maximum number of retries for failed requests. Defaults to 3
        openai_api_key (str, optional): OpenAI API key. Will look for OPENAI_API_KEY env variable if not provided
        claude_api_key (str, optional): Claude API key. Will look for CLAUDE_API_KEY env variable if not provided
        xai_api_key (str, optional): XAI API key. Will look for XAI_API_KEY env variable if not provided
        gemini_api_key (str, optional): Gemini API key. Will look for GEMINI_API_KEY env variable if not provided

    Attributes:
        files: Access to file operations
        fine_tuning: Access to model fine-tuning operations
        prompt_optimization: Access to prompt optimization operations
        documents: Access to document operations
        models: Access to model operations
        datasets: Access to dataset operations
        schemas: Access to schema operations
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
        self.files_datasets = files_datasets.Datasets(client=self)
    def _request(
            self, method: str, endpoint: str, data: Optional[dict[str, Any]] = None, idempotency_key: str | None = None, files: Optional[List[FileTuple]] = None
    ) -> Any:
        """Makes a synchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.)
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload
            idempotency_key (str, optional): Idempotency key for request

        Returns:
            Any: Parsed JSON response

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """

        @backoff.on_exception(backoff.expo, httpx.HTTPStatusError, max_tries=self.max_retries + 1, on_giveup=raise_max_tries_exceeded)
        def wrapped_request() -> Any:
            if method == "GET":
                response = self.client.request(
                    method,
                    self._prepare_url(endpoint),
                    params=data,  # Use data as query params for GET
                    headers=self._get_headers(idempotency_key)
                )
            elif files:  # Handle requests with file uploads
                headers = self._get_headers(idempotency_key)
                headers.pop("Content-Type", None)  # Remove Content-Type if present
                response = self.client.request(
                    method,
                    self._prepare_url(endpoint),
                    params=data,  # Query parameters
                    files=files,  # File data
                    headers=headers
                 )
            else:
                response = self.client.request(
                    method,
                    self._prepare_url(endpoint),
                    json=data,  # Use data as JSON body for non-GET
                    headers=self._get_headers(idempotency_key)
                )

            self._validate_response(response)
            return response.json()

        return wrapped_request()



    def _request_stream(
            self, method: str, endpoint: str, data: Optional[dict[str, Any]] = None, idempotency_key: str | None = None
    ) -> Iterator[Any]:
        """Makes a streaming synchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.) 
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload
            idempotency_key (str, optional): Idempotency key for request
        Returns:
            Iterator[Any]: Generator yielding parsed JSON objects from the stream

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """
        @backoff.on_exception(backoff.expo, httpx.HTTPStatusError, max_tries=self.max_retries + 1, on_giveup=raise_max_tries_exceeded)
        def wrapped_request() -> Iterator[Any]:
            with self.client.stream(method, self._prepare_url(endpoint), json=data, headers=self._get_headers(idempotency_key)) as response_ctx_manager:
                self._validate_response(response_ctx_manager)
                
                for chunk in response_ctx_manager.iter_lines():
                    if not chunk: continue
                    try: yield json.loads(chunk)
                    except Exception: pass
        
        for item in wrapped_request():
            yield item

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
    
    This client provides asynchronous access to all UiForm API resources including files, fine-tuning,
    prompt optimization, documents, models, datasets, and schemas.

    Args:
        api_key (str, optional): UiForm API key. If not provided, will look for UIFORM_API_KEY env variable.
        base_url (str, optional): Base URL for API requests. Defaults to https://api.uiform.com
        timeout (float): Request timeout in seconds. Defaults to 240.0
        max_retries (int): Maximum number of retries for failed requests. Defaults to 3
        openai_api_key (str, optional): OpenAI API key. Will look for OPENAI_API_KEY env variable if not provided
        claude_api_key (str, optional): Claude API key. Will look for CLAUDE_API_KEY env variable if not provided
        xai_api_key (str, optional): XAI API key. Will look for XAI_API_KEY env variable if not provided
        gemini_api_key (str, optional): Gemini API key. Will look for GEMINI_API_KEY env variable if not provided

    Attributes:
        files: Access to asynchronous file operations
        fine_tuning: Access to asynchronous model fine-tuning operations
        prompt_optimization: Access to asynchronous prompt optimization operations
        documents: Access to asynchronous document operations
        models: Access to asynchronous model operations
        datasets: Access to asynchronous dataset operations
        schemas: Access to asynchronous schema operations
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
        self, method: str, endpoint: str, data: Optional[dict[str, Any]] = None, idempotency_key: str | None = None
    ) -> Any:
        """Makes an asynchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.)
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload
            idempotency_key (str, optional): Idempotency key for request
        Returns:
            Any: Parsed JSON response

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """
        @backoff.on_exception(backoff.expo, httpx.HTTPStatusError, max_tries=self.max_retries + 1, on_giveup=raise_max_tries_exceeded)
        async def wrapped_request() -> Any:
            response = await self.client.request(
                    method, self._prepare_url(endpoint), json=data, headers=self._get_headers(idempotency_key)
                )
            self._validate_response(response)

            return response.json()

        return await wrapped_request()
        
    async def _request_stream(
        self, method: str, endpoint: str, data: Optional[dict[str, Any]] = None, idempotency_key: str | None = None
    ) -> AsyncIterator[Any]:
        """Makes a streaming asynchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.)
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload
            idempotency_key (str, optional): Idempotency key for request
        Returns:
            AsyncIterator[Any]: Async generator yielding parsed JSON objects from the stream

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """
        @backoff.on_exception(backoff.expo, httpx.HTTPStatusError, max_tries=self.max_retries + 1, on_giveup=raise_max_tries_exceeded)
        async def wrapped_request() -> AsyncIterator[Any]:
            async with self.client.stream(method, self._prepare_url(endpoint), json=data, headers=self._get_headers(idempotency_key)) as response_ctx_manager:
                self._validate_response(response_ctx_manager)
                async for chunk in response_ctx_manager.aiter_lines():
                    if not chunk: continue
                    try: yield json.loads(chunk)
                    except Exception: pass
        
        async for item in wrapped_request():
            yield item

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
