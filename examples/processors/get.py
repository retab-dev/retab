from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../.env.production")

reclient = Retab()

processor = reclient.processors.get(
    processor_id="proc_t51R0NeVxvWlBFEJJNVAd",
)

print(processor.model_dump_json(indent=2))
