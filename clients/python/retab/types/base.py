from pydantic import BaseModel, ConfigDict


class RetabBaseModel(BaseModel):
    model_config = ConfigDict(
        extra="ignore",
        populate_by_name=True,
        protected_namespaces=(),
    )
