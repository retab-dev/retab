from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../.env.production")

reclient = Retab()

processor = reclient.processors.update(
    processor_id="proc_F0FE8DFqyouQdZXDTWRg0",
    name="Invoice Processor Updated",
)

print(processor.model_dump_json(indent=2))
