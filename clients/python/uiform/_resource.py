from __future__ import annotations
import time
import asyncio
from typing import TYPE_CHECKING

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
