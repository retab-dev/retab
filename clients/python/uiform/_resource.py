from __future__ import annotations
import time
import asyncio
from typing import TYPE_CHECKING, Optional, Callable, Any
from concurrent.futures import ThreadPoolExecutor

if TYPE_CHECKING:
    from .client import UiForm, AsyncUiForm


class SyncAPIResource:
    _client: UiForm

    def __init__(self, client: UiForm) -> None:
        self._client = client

    def _sleep(self, seconds: float) -> None:
        time.sleep(seconds)


class AsyncAPIResource:
    _client: AsyncUiForm

    def __init__(self, client: AsyncUiForm) -> None:
        self._client = client

    async def _sleep(self, seconds: float) -> None:
        await asyncio.sleep(seconds)

class WrappedAsyncAPIResource:
    def __init__(self, client: AsyncUiForm, sync_resource: SyncAPIResource, max_workers: Optional[int] = None) -> None:
        """
        Initialize the WrappedAsyncAPIResource.

        :param sync_resource: An instance of SyncAPIResource to wrap.
        :param max_workers: Optional maximum number of workers for the ThreadPoolExecutor.
        """
        self._client = client
        self._sync_resource = sync_resource
        self._executor = ThreadPoolExecutor(max_workers=max_workers)

    def __getattr__(self, name: str) -> Callable:
        """
        Dynamically handle method calls.

        :param name: Name of the method to call.
        :return: Wrapped method or original method.
        """
        sync_method = getattr(self._sync_resource, name, None)

        if callable(sync_method):
            async def wrapped_method(*args: Any, **kwargs: Any) -> Any:
                loop = asyncio.get_event_loop()
                return await loop.run_in_executor(
                    self._executor, lambda: sync_method(*args, **kwargs)
                )

            return wrapped_method

        raise AttributeError(f"'{type(self).__name__}' object has no attribute '{name}'")