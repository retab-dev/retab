import importlib
import inspect
import pkgutil

from pydantic import BaseModel

import retab.types

import pytest

# Whole module is unit (pure offline; no server/credentials needed).
pytestmark = pytest.mark.unit


def _iter_sdk_pydantic_models() -> list[type[BaseModel]]:
    models: list[type[BaseModel]] = []
    for module_info in pkgutil.walk_packages(retab.types.__path__, f"{retab.types.__name__}."):
        module = importlib.import_module(module_info.name)
        for value in vars(module).values():
            if not inspect.isclass(value):
                continue
            if not issubclass(value, BaseModel):
                continue
            if not value.__module__.startswith("retab.types"):
                continue
            models.append(value)
    return sorted(set(models), key=lambda model: f"{model.__module__}.{model.__qualname__}")


def test_sdk_pydantic_models_ignore_extra_fields() -> None:
    """Response models must declare ``extra="ignore"`` so a newer server adding
    a field doesn't break the SDK.

    Request models are exempt: they describe what the caller sends ON THE WIRE
    and may legitimately use ``extra="forbid"`` to catch typos in caller-
    supplied kwargs. The convention is ``<Noun>Request`` /
    ``<Verb><Noun>Request`` (e.g. ``ExtractionRequest``, ``CreateJobRequest``,
    ``UpdateEditTemplateRequest``).
    """
    models = _iter_sdk_pydantic_models()

    assert models
    violators = [f"{model.__module__}.{model.__qualname__}" for model in models if not model.__qualname__.endswith("Request") and model.model_config.get("extra") != "ignore"]
    assert violators == [], violators
