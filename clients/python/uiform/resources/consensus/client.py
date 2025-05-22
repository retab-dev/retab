from typing import Any, Dict, List, Literal, Optional

from pydantic import BaseModel, Field

from ..._resource import AsyncAPIResource, SyncAPIResource
from ...types.standards import PreparedRequest


class BaseConsensusMixin:
    def _prepare_compare_extractions(
        self,
        dict1: Dict[str, Any],
        dict2: Dict[str, Any],
        metric: str = "levenshtein_similarity",
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
        data = {
            "dict1": dict1,
            "dict2": dict2,
            "metric": metric,
        }
        
        return PreparedRequest(
            method="POST", 
            url="/v1/consensus/distances", 
            data=data, 
            idempotency_key=idempotency_key
        )

    def _prepare_reconcile(
        self,
        list_dicts: List[Dict[str, Any]],
        reference_schema: Optional[Dict[str, Any]] = None,
        mode: Literal["direct", "aligned"] = "direct",
        min_support_ratio: float = 0.51,
        idempotency_key: str | None = None,
    ) -> PreparedRequest:
        data = {
            "list_dicts": list_dicts,
            "mode": mode,
            "min_support_ratio": min_support_ratio,
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

    def compare_extractions(
        self,
        dict1: Dict[str, Any],
        dict2: Dict[str, Any],
        metric: str = "levenshtein_similarity",
        idempotency_key: str | None = None,
    ) -> Dict[str, Any]:
        """
        Compare two dictionaries using the specified similarity metric.
        
        Args:
            dict1: First dictionary to compare
            dict2: Second dictionary to compare
            metric: Similarity metric to use (default: "levenshtein_similarity")
            idempotency_key: Optional idempotency key for the request
            
        Returns:
            Dict containing comparison results with similarity scores for each field
            
        Raises:
            UiformAPIError: If the API request fails
        """
        request = self._prepare_compare_extractions(dict1, dict2, metric, idempotency_key)
        response = self._client._prepared_request(request)
        return response
    
    def reconcile(
        self,
        list_dicts: List[Dict[str, Any]],
        reference_schema: Optional[Dict[str, Any]] = None,
        mode: Literal["direct", "aligned"] = "direct",
        min_support_ratio: float = 0.51,
        idempotency_key: str | None = None,
    ) -> Dict[str, Any]:
        """
        Reconcile multiple dictionaries to produce a single unified consensus dictionary.
        
        Args:
            list_dicts: List of dictionaries to reconcile
            reference_schema: Optional schema to validate dictionaries against
            mode: Mode for consensus computation ("direct" or "aligned")
            min_support_ratio: Minimum support ratio for reference elements
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
            min_support_ratio,
            idempotency_key
        )
        response = self._client._prepared_request(request)
        return response


class AsyncConsensus(AsyncAPIResource, BaseConsensusMixin):
    """Consensus API wrapper for asynchronous operations"""

    async def compare_extractions(
        self,
        dict1: Dict[str, Any],
        dict2: Dict[str, Any],
        metric: str = "levenshtein_similarity",
        idempotency_key: str | None = None,
    ) -> Dict[str, Any]:
        """
        Compare two dictionaries using the specified similarity metric asynchronously.
        
        Args:
            dict1: First dictionary to compare
            dict2: Second dictionary to compare
            metric: Similarity metric to use (default: "levenshtein_similarity")
            idempotency_key: Optional idempotency key for the request
            
        Returns:
            Dict containing comparison results with similarity scores for each field
            
        Raises:
            UiformAPIError: If the API request fails
        """
        request = self._prepare_compare_extractions(dict1, dict2, metric, idempotency_key)
        response = await self._client._prepared_request(request)
        return response
    
    async def reconcile(
        self,
        list_dicts: List[Dict[str, Any]],
        reference_schema: Optional[Dict[str, Any]] = None,
        mode: Literal["direct", "aligned"] = "direct",
        min_support_ratio: float = 0.51,
        idempotency_key: str | None = None,
    ) -> Dict[str, Any]:
        """
        Reconcile multiple dictionaries to produce a single unified consensus dictionary asynchronously.
        
        Args:
            list_dicts: List of dictionaries to reconcile
            reference_schema: Optional schema to validate dictionaries against
            mode: Mode for consensus computation ("direct" or "aligned")
            min_support_ratio: Minimum support ratio for reference elements
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
            min_support_ratio,
            idempotency_key
        )
        response = await self._client._prepared_request(request)
        return response
