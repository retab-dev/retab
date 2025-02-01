from typing import IO, Any, Optional
from pathlib import Path
import time
from pydantic import BaseModel
import PIL.Image
from typing import Type, Optional
from io import IOBase

from ..types.schemas.generate import GenerateSchemaRequest
from ..types.schemas.object import Schema
from ..types.schemas.promptify import PromptifyRequest
from ..types.modalities import Modality
from .._resource import SyncAPIResource, AsyncAPIResource

from .._utils.json_schema import load_json_schema
from .._utils.mime import prepare_mime_document_list
from .._utils.ai_model import assert_valid_model_schema_generation

from typing import List

import datetime
from ..types.usage import Amount

total_cost = 0.0

class Usage(SyncAPIResource):

    def total(self, start_date: Optional[datetime.datetime] = None, end_date: Optional[datetime.datetime] = None) -> Amount:
        """Get the total usage cost for a mailbox within an optional date range.

        Returns:
            Amount: The total usage cost
        """
        return Amount(value=total_cost, currency="USD")
    
    def mailbox(self, email: str, start_date: Optional[datetime.datetime] = None, end_date: Optional[datetime.datetime] = None) -> Amount:
        """Get the total usage cost for a mailbox within an optional date range.

        Args:
            email: The email address of the mailbox
            start_date: Start date for usage calculation
            end_date: End date for usage calculation

        Returns:
            Amount: The total usage cost
        """
        params = {}
        if start_date:
            params["start_date"] = start_date.isoformat()
        if end_date:
            params["end_date"] = end_date.isoformat()

        response = self._client._request(
            "GET", 
            f"/v1/automations/mailboxes/{email}/usage",
            params=params
        )
        return Amount.model_validate(response)

    def link(self, link_id: str, start_date: Optional[datetime.datetime] = None, end_date: Optional[datetime.datetime] = None) -> Amount:
        """Get the total usage cost for a link within an optional date range.

        Args:
            link_id: The ID of the link
            start_date: Start date for usage calculation
            end_date: End date for usage calculation

        Returns:
            Amount: The total usage cost
        """
        params = {}
        if start_date:
            params["start_date"] = start_date.isoformat()
        if end_date:
            params["end_date"] = end_date.isoformat()

        response = self._client._request(
            "GET",
            f"/v1/automations/links/{link_id}/usage",
            params=params
        )
        return Amount.model_validate(response)
    
    def schema(self, schema_id: str, start_date: Optional[datetime.datetime] = None, end_date: Optional[datetime.datetime] = None) -> Amount:
        """Get the total usage cost for a schema within an optional date range.

        Args:
            schema_id: The ID of the schema
            start_date: Start date for usage calculation
            end_date: End date for usage calculation

        Returns:
            Amount: The total usage cost
        """
        params = {}
        if start_date:
            params["start_date"] = start_date.isoformat()
        if end_date:
            params["end_date"] = end_date.isoformat()

        response = self._client._request(
            "GET",
            f"/v1/schemas/{schema_id}/usage",
            params=params
        )
        return Amount.model_validate(response)


    def schema_data(self, schema_data_id: str, start_date: Optional[datetime.datetime] = None, end_date: Optional[datetime.datetime] = None) -> Amount:
        """Get the total usage cost for a schema within an optional date range.

            Args:
                schema_id: The ID of the schema
                start_date: Start date for usage calculation
                end_date: End date for usage calculation

            Returns:
                Amount: The total usage cost
        """
        params = {}
        if start_date:
            params["start_date"] = start_date.isoformat()
        if end_date:
            params["end_date"] = end_date.isoformat()

        response = self._client._request(
            "GET",
            f"/v1/schemas/{schema_data_id}/usage_data",
            params=params
        )
        return Amount.model_validate(response)
