from enum import Enum
from pydantic import BaseModel, constr

import outlines
from uiform import Schema

class Weapon(str, Enum):
    sword = "sword"
    axe = "axe"
    mace = "mace"
    spear = "spear"
    bow = "bow"
    crossbow = "crossbow"


class Armor(str, Enum):
    leather = "leather"
    chainmail = "chainmail"
    plate = "plate"


class Character(BaseModel):
    name: constr(max_length=10)
    age: int
    armor: Armor
    weapon: Weapon
    strength: int


model = outlines.models.transformers("microsoft/Phi-3-mini-4k-instruct")

# Construct structured sequence generator
generator = outlines.generate.json(model, Character)


schema_obj = Schema(
    pydantic_model=Character
)

character = generator("Give me a character description")

print(repr(character))

extraction = schema_obj.pydantic_model.model_validate(
    character.model_dump()
)

print("Result:",extraction)
