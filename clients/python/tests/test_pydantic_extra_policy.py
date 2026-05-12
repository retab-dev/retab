import importlib
import inspect
import pkgutil

from pydantic import BaseModel

import retab.types
from retab.types.base import RetabBaseModel


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
    models = _iter_sdk_pydantic_models()

    assert models
    assert [
        f"{model.__module__}.{model.__qualname__}"
        for model in models
        if not issubclass(model, RetabBaseModel) or model.model_config.get("extra") != "ignore"
    ] == []
