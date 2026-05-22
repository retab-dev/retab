from pydantic import BaseModel, ConfigDict


class RetabBaseModel(BaseModel):
    model_config = ConfigDict(
        extra="ignore",
        # Generated models alias fields that would shadow Pydantic reserved
        # names (e.g. `schema` → Python `schema_`); allow callers to pass
        # either form when constructing models.
        populate_by_name=True,
        # Silence `Field name "model_X" shadows attribute in parent` warnings
        # for any IR field that legitimately starts with `model_`.
        protected_namespaces=(),
    )
