from typing import Any, Dict, List, Literal, Optional

from pydantic import BaseModel, Field

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.standards import PreparedRequest
from .completions import AsyncCompletions, Completions
from .responses import AsyncResponses, Responses
from ...types.consensus import ReconciliationResponse

class BaseConsensusMixin:

    def _prepare_reconcile(
        self,
        list_dicts: List[Dict[str, Any]],
        reference_schema: Optional[Dict[str, Any]] = None,
        mode: Literal["direct", "aligned"] = "direct",
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
        data = {
            "list_dicts": list_dicts,
            "mode": mode,
        }
        
        if reference_schema is not None:
            data["reference_schema"] = reference_schema
            
        return PreparedRequest(
            method="POST", 
            url="/v1/consensus/reconcile", 
            data=data, 
            idempotency_key=idempotency_key
        )

class Consensus(SyncAPIResource, BaseConsensusMixin):
    """Consensus API wrapper for synchronous operations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.completions = Completions(client=client)
        self.responses = Responses(client=client)

    def reconcile(
        self,
        list_dicts: List[Dict[str, Any]],
        reference_schema: Optional[Dict[str, Any]] = None,
        mode: Literal["direct", "aligned"] = "direct",
        idempotency_key: str | None = None,
    ) -> ReconciliationResponse:
        """
        Reconcile multiple dictionaries to produce a single unified consensus dictionary.
        
        Args:
            list_dicts: List of dictionaries to reconcile
            reference_schema: Optional schema to validate dictionaries against
            mode: Mode for consensus computation ("direct" or "aligned")
            idempotency_key: Optional idempotency key for the request
            
        Returns:
            Dict containing the consensus dictionary and consensus likelihoods
            
        Raises:
            UiformAPIError: If the API request fails
        """
        request = self._prepare_reconcile(
            list_dicts, 
            reference_schema, 
            mode, 
            idempotency_key
        )
        response = self._client._prepared_request(request)
        return ReconciliationResponse.model_validate(response)


class AsyncConsensus(AsyncAPIResource, BaseConsensusMixin):
    """Consensus API wrapper for asynchronous operations"""

    def __init__(self, client: Any) -> None:
        super().__init__(client=client)
        self.completions = AsyncCompletions(client=client)
        self.responses = AsyncResponses(client=client)

    
    async def reconcile(
        self,
        list_dicts: List[Dict[str, Any]],
        reference_schema: Optional[Dict[str, Any]] = None,
        mode: Literal["direct", "aligned"] = "direct",
        idempotency_key: str | None = None,
    ) -> ReconciliationResponse:
        """
        Reconcile multiple dictionaries to produce a single unified consensus dictionary asynchronously.
        
        Args:
            list_dicts: List of dictionaries to reconcile
            reference_schema: Optional schema to validate dictionaries against
            mode: Mode for consensus computation ("direct" or "aligned")
            idempotency_key: Optional idempotency key for the request
            
        Returns:
            Dict containing the consensus dictionary and consensus likelihoods
            
        Raises:
            UiformAPIError: If the API request fails
        """
        request = self._prepare_reconcile(
            list_dicts, 
            reference_schema, 
            mode, 
            idempotency_key
        )
        response = await self._client._prepared_request(request)

        return ReconciliationResponse.model_validate(response)
