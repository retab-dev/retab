import json
import os
from types import TracebackType
from typing import Any, AsyncIterator, BinaryIO, Iterator, List, Optional

import backoff
import backoff.types
import httpx
from pydantic_core import PydanticUndefined

from .resources import deployments, consensus, documents, evals, files, finetuning, models, schemas, secrets, usage
from .types.standards import PreparedRequest
import truststore


class MaxRetriesExceeded(Exception):
    pass


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

    Raises:
        ValueError: If no API key is provided through arguments or environment variables
    """

    #   claude_api_key (str, optional): Claude API key. Will look for CLAUDE_API_KEY env variable if not provided
    #   xai_api_key (str, optional): XAI API key. Will look for XAI_API_KEY env variable if not provided
    #   gemini_api_key (str, optional): Gemini API key. Will look for GEMINI_API_KEY env variable if not provided

    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: Optional[str] = None,
        timeout: float = 240.0,
        max_retries: int = 3,
        openai_api_key: Optional[str] = PydanticUndefined,  # type: ignore[assignment]
        gemini_api_key: Optional[str] = PydanticUndefined,  # type: ignore[assignment]
        # claude_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
        xai_api_key: Optional[str] = PydanticUndefined,  # type: ignore[assignment]
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


        truststore.inject_into_ssl()
        self.api_key = api_key
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.max_retries = max_retries
        self.headers = {
            "Api-Key": self.api_key,
            "Content-Type": "application/json",
        }

        # Only check environment variables if the value is PydanticUndefined
        if openai_api_key is PydanticUndefined:
            openai_api_key = os.environ.get("OPENAI_API_KEY")
        
        # if claude_api_key is PydanticUndefined:
        #    claude_api_key = os.environ.get("CLAUDE_API_KEY")
        
        # if xai_api_key is PydanticUndefined:
        #    xai_api_key = os.environ.get("XAI_API_KEY")
        
        if gemini_api_key is PydanticUndefined:
            gemini_api_key = os.environ.get("GEMINI_API_KEY")

        # Only add headers if the values are actual strings (not None or PydanticUndefined)
        if openai_api_key and openai_api_key is not PydanticUndefined:
            self.headers["OpenAI-Api-Key"] = openai_api_key
        
        # if claude_api_key and claude_api_key is not PydanticUndefined:
        #    self.headers["Anthropic-Api-Key"] = claude_api_key
        
        if xai_api_key and xai_api_key is not PydanticUndefined:
            self.headers["XAI-Api-Key"] = xai_api_key
        
        if gemini_api_key and gemini_api_key is not PydanticUndefined:
            self.headers["Gemini-Api-Key"] = gemini_api_key

    def _prepare_url(self, endpoint: str) -> str:
        return f"{self.base_url}/{endpoint.lstrip('/')}"

    def _validate_response(self, response_object: httpx.Response) -> None:
        if response_object.status_code >= 500:
            response_object.raise_for_status()
        elif response_object.status_code == 422:
            raise RuntimeError(f"Validation error (422): {response_object.text}")
        elif not response_object.is_success:
            raise RuntimeError(f"Request failed ({response_object.status_code}): {response_object.text}")

    def _get_headers(self, idempotency_key: str | None = None) -> dict[str, Any]:
        headers = self.headers.copy()
        if idempotency_key:
            headers["Idempotency-Key"] = idempotency_key
        return headers

    def _parse_response(self, response: httpx.Response) -> Any:
        """Parse response based on content-type.
        
        Returns:
            Any: Parsed JSON object for JSON responses, raw text string for text responses
        """
        content_type = response.headers.get("content-type", "")
        
        # Check if it's a JSON response
        if "application/json" in content_type or "application/stream+json" in content_type:
            return response.json()
        # Check if it's a text response
        elif "text/plain" in content_type or "text/" in content_type:
            return response.text
        else:
            # Default to JSON parsing for backwards compatibility
            try:
                return response.json()
            except Exception:
                # If JSON parsing fails, return as text
                return response.text


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
        responses: Access to responses API (OpenAI Responses API compatible interface)
    """

    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: Optional[str] = None,
        timeout: float = 240.0,
        max_retries: int = 3,
        openai_api_key: Optional[str] = PydanticUndefined,  # type: ignore[assignment]
        gemini_api_key: Optional[str] = PydanticUndefined,  # type: ignore[assignment]
        # claude_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
        # xai_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
    ) -> None:
        super().__init__(
            api_key=api_key,
            base_url=base_url,
            timeout=timeout,
            max_retries=max_retries,
            openai_api_key=openai_api_key,
            gemini_api_key=gemini_api_key,
            # claude_api_key=claude_api_key,
            # xai_api_key=xai_api_key,
        )

        self.client = httpx.Client(timeout=self.timeout)
        self.evals = evals.Evals(client=self)
        self.files = files.Files(client=self)
        self.fine_tuning = finetuning.FineTuning(client=self)
        # self.prompt_optimization = prompt_optimization.PromptOptimization(client=self)
        self.documents = documents.Documents(client=self)
        self.models = models.Models(client=self)
        self.schemas = schemas.Schemas(client=self)
        self.deployments = deployments.Deployments(client=self)
        self.secrets = secrets.Secrets(client=self)
        self.usage = usage.Usage(client=self)
        self.consensus = consensus.Consensus(client=self)

    def _request(
        self,
        method: str,
        endpoint: str,
        data: Optional[dict[str, Any]] = None,
        params: Optional[dict[str, Any]] = None,
        idempotency_key: str | None = None,
        raise_for_status: bool = False,
    ) -> Any:
        """Makes a synchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.)
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload
            params (Optional[dict]): Query parameters
            idempotency_key (str, optional): Idempotency key for request

        Returns:
            Any: Parsed JSON response or raw text string depending on response content-type

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """

        def raw_request() -> Any:
            response = self.client.request(method, json=data, url=self._prepare_url(endpoint), params=params, headers=self._get_headers(idempotency_key))

            self._validate_response(response)

            return self._parse_response(response)

        @backoff.on_exception(backoff.expo, httpx.HTTPStatusError, max_tries=self.max_retries + 1, on_giveup=raise_max_tries_exceeded)
        def wrapped_request() -> Any:
            return raw_request()

        if raise_for_status:
            # If raise_for_status is True, we want to raise an exception if the request fails, not retry...
            return raw_request()
        else:
            return wrapped_request()

    def _request_stream(
        self,
        method: str,
        endpoint: str,
        data: Optional[dict[str, Any]] = None,
        params: Optional[dict[str, Any]] = None,
        idempotency_key: str | None = None,
        raise_for_status: bool = False,
    ) -> Iterator[Any]:
        """Makes a streaming synchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.)
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload
            params (Optional[dict]): Query parameters
            idempotency_key (str, optional): Idempotency key for request
        Returns:
            Iterator[Any]: Generator yielding parsed JSON objects or raw text strings from the stream

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """

        def raw_request() -> Iterator[Any]:
            with self.client.stream(method, self._prepare_url(endpoint), json=data, params=params, headers=self._get_headers(idempotency_key)) as response_ctx_manager:
                self._validate_response(response_ctx_manager)
                
                content_type = response_ctx_manager.headers.get("content-type", "")
                is_json_stream = "application/json" in content_type or "application/stream+json" in content_type
                is_text_stream = "text/plain" in content_type or ("text/" in content_type and not is_json_stream)

                for chunk in response_ctx_manager.iter_lines():
                    if not chunk:
                        continue
                    
                    if is_json_stream:
                        try:
                            yield json.loads(chunk)
                        except Exception:
                            pass
                    elif is_text_stream:
                        yield chunk
                    else:
                        # Default behavior: try JSON first, fall back to text
                        try:
                            yield json.loads(chunk)
                        except Exception:
                            yield chunk

        @backoff.on_exception(backoff.expo, httpx.HTTPStatusError, max_tries=self.max_retries + 1, on_giveup=raise_max_tries_exceeded)
        def wrapped_request() -> Iterator[Any]:
            for item in raw_request():
                yield item

        iterator_ = raw_request() if raise_for_status else wrapped_request()

        for item in iterator_:
            yield item

    # Simplified request methods using standard PreparedRequest object
    def _prepared_request(self, request: PreparedRequest) -> Any:
        return self._request(
            request.method, request.url, data=request.data, params=request.params, idempotency_key=request.idempotency_key, raise_for_status=request.raise_for_status
        )

    def _prepared_request_stream(self, request: PreparedRequest) -> Iterator[Any]:
        for item in self._request_stream(request.method, request.url, data=request.data, params=request.params, idempotency_key=request.idempotency_key):
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
        responses: Access to responses API (OpenAI Responses API compatible interface)
    """

    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: Optional[str] = None,
        timeout: float = 240.0,
        max_retries: int = 3,
        openai_api_key: Optional[str] = PydanticUndefined,  # type: ignore[assignment]
        gemini_api_key: Optional[str] = PydanticUndefined,  # type: ignore[assignment]
        # claude_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
        # xai_api_key: Optional[str] = PydanticUndefined,   # type: ignore[assignment]
    ) -> None:
        super().__init__(
            api_key=api_key,
            base_url=base_url,
            timeout=timeout,
            max_retries=max_retries,
            openai_api_key=openai_api_key,
            gemini_api_key=gemini_api_key,
            # claude_api_key=claude_api_key,
            # xai_api_key=xai_api_key,
        )

        self.client = httpx.AsyncClient(timeout=self.timeout)

        self.evals = evals.AsyncEvals(client=self)
        self.files = files.AsyncFiles(client=self)
        self.fine_tuning = finetuning.AsyncFineTuning(client=self)
        # self.prompt_optimization = prompt_optimization.AsyncPromptOptimization(client=self)
        self.documents = documents.AsyncDocuments(client=self)
        self.models = models.AsyncModels(client=self)
        self.schemas = schemas.AsyncSchemas(client=self)
        self.deployments = deployments.AsyncDeployments(client=self)
        self.secrets = secrets.AsyncSecrets(client=self)
        self.usage = usage.AsyncUsage(client=self)
        self.consensus = consensus.AsyncConsensus(client=self)

    def _parse_response(self, response: httpx.Response) -> Any:
        """Parse response based on content-type.
        
        Returns:
            Any: Parsed JSON object for JSON responses, raw text string for text responses
        """
        content_type = response.headers.get("content-type", "")
        
        # Check if it's a JSON response
        if "application/json" in content_type or "application/stream+json" in content_type:
            return response.json()
        # Check if it's a text response
        elif "text/plain" in content_type or "text/" in content_type:
            return response.text
        else:
            # Default to JSON parsing for backwards compatibility
            try:
                return response.json()
            except Exception:
                # If JSON parsing fails, return as text
                return response.text

    async def _request(
        self,
        method: str,
        endpoint: str,
        data: Optional[dict[str, Any]] = None,
        params: Optional[dict[str, Any]] = None,
        idempotency_key: str | None = None,
        raise_for_status: bool = False,
    ) -> Any:
        """Makes an asynchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.)
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload
            params (Optional[dict]): Query parameters
            idempotency_key (str, optional): Idempotency key for request
        Returns:
            Any: Parsed JSON response or raw text string depending on response content-type

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """

        async def raw_request() -> Any:
            response = await self.client.request(method, self._prepare_url(endpoint), json=data, params=params, headers=self._get_headers(idempotency_key))
            self._validate_response(response)
            return self._parse_response(response)

        @backoff.on_exception(backoff.expo, httpx.HTTPStatusError, max_tries=self.max_retries + 1, on_giveup=raise_max_tries_exceeded)
        async def wrapped_request() -> Any:
            return await raw_request()

        if raise_for_status:
            return await raw_request()
        else:
            return await wrapped_request()

    async def _request_stream(
        self,
        method: str,
        endpoint: str,
        data: Optional[dict[str, Any]] = None,
        params: Optional[dict[str, Any]] = None,
        idempotency_key: str | None = None,
        raise_for_status: bool = False,
    ) -> AsyncIterator[Any]:
        """Makes a streaming asynchronous HTTP request to the API.

        Args:
            method (str): HTTP method (GET, POST, etc.)
            endpoint (str): API endpoint path
            data (Optional[dict]): Request payload
            params (Optional[dict]): Query parameters
            idempotency_key (str, optional): Idempotency key for request
        Returns:
            AsyncIterator[Any]: Async generator yielding parsed JSON objects or raw text strings from the stream

        Raises:
            RuntimeError: If request fails after max retries or validation error occurs
        """

        async def raw_request() -> AsyncIterator[Any]:
            async with self.client.stream(method, self._prepare_url(endpoint), json=data, params=params, headers=self._get_headers(idempotency_key)) as response_ctx_manager:
                self._validate_response(response_ctx_manager)
                
                content_type = response_ctx_manager.headers.get("content-type", "")
                is_json_stream = "application/json" in content_type or "application/stream+json" in content_type
                is_text_stream = "text/plain" in content_type or ("text/" in content_type and not is_json_stream)
                
                async for chunk in response_ctx_manager.aiter_lines():
                    if not chunk:
                        continue
                    
                    if is_json_stream:
                        try:
                            yield json.loads(chunk)
                        except Exception:
                            pass
                    elif is_text_stream:
                        yield chunk
                    else:
                        # Default behavior: try JSON first, fall back to text
                        try:
                            yield json.loads(chunk)
                        except Exception:
                            yield chunk

        @backoff.on_exception(backoff.expo, httpx.HTTPStatusError, max_tries=self.max_retries + 1, on_giveup=raise_max_tries_exceeded)
        async def wrapped_request() -> AsyncIterator[Any]:
            async for item in raw_request():
                yield item

        async_iterator_ = raw_request() if raise_for_status else wrapped_request()

        async for item in async_iterator_:
            yield item

    async def _prepared_request(self, request: PreparedRequest) -> Any:
        return await self._request(
            request.method, request.url, data=request.data, params=request.params, idempotency_key=request.idempotency_key, raise_for_status=request.raise_for_status
        )

    async def _prepared_request_stream(self, request: PreparedRequest) -> AsyncIterator[Any]:
        async for item in self._request_stream(
            request.method, request.url, data=request.data, params=request.params, idempotency_key=request.idempotency_key, raise_for_status=request.raise_for_status
        ):
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
