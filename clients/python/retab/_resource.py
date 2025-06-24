from __future__ import annotations

import asyncio
import time
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .client import AsyncRetab, Retab


class SyncAPIResource:
    _client: Retab

    def __init__(self, client: Retab) -> None:
        self._client = client

    def _sleep(self, seconds: float) -> None:
        time.sleep(seconds)


class AsyncAPIResource:
    _client: AsyncRetab

    def __init__(self, client: AsyncRetab) -> None:
        self._client = client

    async def _sleep(self, seconds: float) -> None:
        await asyncio.sleep(seconds)
