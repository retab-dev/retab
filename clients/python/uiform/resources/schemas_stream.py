from io import IOBase
from pathlib import Path
from typing import Any, AsyncGenerator, Generator, Sequence

import PIL.Image
from pydantic import BaseModel

from .._resource import AsyncAPIResource, SyncAPIResource
from .._utils.ai_models import assert_valid_model_schema_generation
from .._utils.json_schema import load_json_schema, unflatten_dict
from .._utils.mime import prepare_mime_document_list
from .._utils.stream_context_managers import as_async_context_manager, as_context_manager
from ..types.modalities import Modality
from ..types.schemas.object import PartialSchema, PartialSchemaChunk, Schema
from ..types.schemas.promptify import PromptifyRequest
from ..types.standards import PreparedRequest


class SchemasStreamMixin:
    def prepare_promptify(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: Sequence[Path | str | bytes | IOBase | PIL.Image.Image],
        model: str,
        temperature: float = 0,
        modality: Modality = "native",
        stream: bool = False,
    ) -> PreparedRequest:
        assert_valid_model_schema_generation(model)
        mime_documents = prepare_mime_document_list(documents)
        loaded_json_schema = load_json_schema(json_schema)
        data = {
            "json_schema": loaded_json_schema,
            "documents": [doc.model_dump() for doc in mime_documents],
            "model": model,
            "temperature": temperature,
            "modality": modality,
            "stream": stream,
        }
        PromptifyRequest.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/schemas/promptify", data=data)

    def prepare_generate(
        self,
        documents: Sequence[Path | str | bytes | IOBase | PIL.Image.Image],
        model: str = "gpt-4o-2024-11-20",
        temperature: float = 0,
        modality: Modality = "native",
        flat: bool = False,
        stream: bool = False,
    ) -> PreparedRequest:
        assert_valid_model_schema_generation(model)
        mime_documents = prepare_mime_document_list(documents)
        data = {"documents": [doc.model_dump() for doc in mime_documents], "model": model, "temperature": temperature, "modality": modality, "flat": flat, "stream": stream}
        return PreparedRequest(method="POST", url="/v1/schemas/generate", data=data)

    def prepare_system_prompt(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: Sequence[Path | str | bytes | IOBase | PIL.Image.Image],
        model: str = "gpt-4o-2024-11-20",
        temperature: float = 0,
        modality: Modality = "native",
        stream: bool = False,
    ) -> PreparedRequest:
        assert_valid_model_schema_generation(model)
        mime_documents = prepare_mime_document_list(documents)
        loaded_json_schema = load_json_schema(json_schema)
        data = {
            "json_schema": loaded_json_schema,
            "documents": [doc.model_dump() for doc in mime_documents],
            "model": model,
            "temperature": temperature,
            "modality": modality,
            "stream": stream,
        }
        PromptifyRequest.model_validate(data)
        return PreparedRequest(method="POST", url="/v1/schemas/system_prompt_endpoint", data=data)


class SchemasStream(SyncAPIResource, SchemasStreamMixin):
    """Streaming methods for Schemas API"""

    @as_context_manager
    def promptify_stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: Sequence[Path | str | bytes | IOBase | PIL.Image.Image],
        model: str,
        temperature: float = 0,
        modality: Modality = "native",
    ) -> Generator[PartialSchema | Schema, None, None]:
        prepared_request = self.prepare_promptify(json_schema, documents, model, temperature, modality, stream=True)
        chunk_json: Any = None
        flat_json_schema = {}

        for chunk_json in self._client._prepared_request_stream(prepared_request):
            if not chunk_json:
                continue

            schema_chunk = PartialSchemaChunk.model_validate(chunk_json)

            # Accumulate the delta updates
            flat_json_schema = {**flat_json_schema, **schema_chunk.delta_json_schema_flat}

            # Unflatten the schema to get proper nested structure
            unflattened_json_schema = unflatten_dict(flat_json_schema)

            # Yield a PartialSchema with the current accumulated state
            yield PartialSchema(json_schema=unflattened_json_schema)

        # The last chunk should yield the full schema object
        if chunk_json:
            final_schema = Schema(json_schema=unflatten_dict(flat_json_schema))
            yield final_schema

    @as_context_manager
    def generate_stream(
        self,
        documents: Sequence[Path | str | bytes | IOBase | PIL.Image.Image],
        model: str = "gpt-4o-2024-11-20",
        temperature: float = 0.0,
        modality: Modality = "native",
        flat: bool = False,
    ) -> Generator[PartialSchema | Schema, None, None]:
        prepared_request = self.prepare_generate(documents, model, temperature, modality, flat, stream=True)
        chunk_json: Any = None
        flat_json_schema = {}

        for chunk_json in self._client._prepared_request_stream(prepared_request):
            if not chunk_json:
                continue

            schema_chunk = PartialSchemaChunk.model_validate(chunk_json)

            # Accumulate the delta updates
            flat_json_schema = {**flat_json_schema, **schema_chunk.delta_json_schema_flat}

            # Unflatten the schema to get proper nested structure
            json_schema = unflatten_dict(flat_json_schema)

            # Yield a PartialSchema with the current accumulated state
            yield PartialSchema(json_schema=json_schema)

        # The last chunk should yield the full schema object
        if chunk_json:
            final_schema = Schema(json_schema=unflatten_dict(flat_json_schema))
            yield final_schema

    @as_context_manager
    def system_prompt_stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: Sequence[Path | str | bytes | IOBase | PIL.Image.Image],
        model: str = "gpt-4o-2024-11-20",
        temperature: float = 0,
        modality: Modality = "native",
    ) -> Generator[str, None, None]:
        prepared_request = self.prepare_system_prompt(json_schema, documents, model, temperature, modality, stream=True)
        for chunk in self._client._prepared_request_stream(prepared_request):
            yield chunk


class AsyncSchemasStream(AsyncAPIResource, SchemasStreamMixin):
    """Async streaming methods for Schemas API"""

    @as_async_context_manager
    async def generate_stream(
        self,
        documents: Sequence[Path | str | bytes | IOBase | PIL.Image.Image],
        model: str = "gpt-4o-2024-11-20",
        temperature: float = 0.0,
        modality: Modality = "native",
        flat: bool = False,
    ) -> AsyncGenerator[PartialSchema | Schema, None]:
        prepared_request = self.prepare_generate(documents, model, temperature, modality, flat, stream=True)
        chunk_json: Any = None
        flat_json_schema = {}

        async for chunk_json in self._client._prepared_request_stream(prepared_request):
            if not chunk_json:
                continue

            schema_chunk = PartialSchemaChunk.model_validate(chunk_json)

            # Accumulate the delta updates
            flat_json_schema = {**flat_json_schema, **schema_chunk.delta_json_schema_flat}

            # Unflatten the schema to get proper nested structure
            json_schema = unflatten_dict(flat_json_schema)

            # Yield a PartialSchema with the current accumulated state
            yield PartialSchema(json_schema=json_schema)

        # The last chunk should yield the full schema object
        if chunk_json:
            final_schema = Schema(json_schema=unflatten_dict(flat_json_schema))
            yield final_schema

    @as_async_context_manager
    async def promptify_stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: Sequence[Path | str | bytes | IOBase | PIL.Image.Image],
        model: str,
        temperature: float = 0,
        modality: Modality = "native",
    ) -> AsyncGenerator[PartialSchema | Schema, None]:
        prepared_request = self.prepare_promptify(json_schema, documents, model, temperature, modality, stream=True)
        chunk_json: Any = None
        flat_json_schema = {}

        async for chunk_json in self._client._prepared_request_stream(prepared_request):
            if not chunk_json:
                continue

            schema_chunk = PartialSchemaChunk.model_validate(chunk_json)

            # Accumulate the delta updates
            flat_json_schema = {**flat_json_schema, **schema_chunk.delta_json_schema_flat}

            # Unflatten the schema to get proper nested structure
            unflattened_json_schema = unflatten_dict(flat_json_schema)

            # Yield a PartialSchema with the current accumulated state
            yield PartialSchema(json_schema=unflattened_json_schema)

        # The last chunk should yield the full schema object
        if chunk_json:
            final_schema = Schema(json_schema=unflatten_dict(flat_json_schema))
            yield final_schema

    @as_async_context_manager
    async def system_prompt_stream(
        self,
        json_schema: dict[str, Any] | Path | str,
        documents: Sequence[Path | str | bytes | IOBase | PIL.Image.Image],
        model: str = "gpt-4o-2024-11-20",
        temperature: float = 0,
        modality: Modality = "native",
    ) -> AsyncGenerator[str, None]:
        prepared_request = self.prepare_system_prompt(json_schema, documents, model, temperature, modality, stream=True)
        async for chunk in self._client._prepared_request_stream(prepared_request):
            yield chunk
