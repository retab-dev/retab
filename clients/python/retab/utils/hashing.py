import base64
import hashlib
import json

from fastapi.encoders import jsonable_encoder

# ************* Generalistic utils *************


def generate_blake2b_hash_from_bytes(bytes_: bytes) -> str:
    return hashlib.blake2b(bytes_, digest_size=8).hexdigest()


def generate_blake2b_hash_from_base64(base64_string: str) -> str:
    return generate_blake2b_hash_from_bytes(base64.b64decode(base64_string))


def generate_blake2b_hash_from_string(input_string: str) -> str:
    return generate_blake2b_hash_from_bytes(input_string.encode("utf-8"))


def generate_blake2b_hash_from_dict(input_dict: dict) -> str:
    jsonable_dict = jsonable_encoder(input_dict)
    return generate_blake2b_hash_from_string(json.dumps(jsonable_dict, sort_keys=True).strip())
