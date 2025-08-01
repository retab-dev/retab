from retab import Retab
from dotenv import load_dotenv

assert load_dotenv("../../.env.production")

reclient = Retab()

processor = reclient.processors.create(
    name="Invoice Processor",
    json_schema="../../assets/code/invoice_schema.json",
    model="gpt-4o-mini",
    modality="native",
    temperature=0.1,
    reasoning_effort="medium",
    n_consensus=3,  # Enable consensus with 3 parallel runs
    image_resolution_dpi=150,
)

completion = reclient.processors.submit(processor_id=processor.id, document="../../assets/docs/invoice.jpeg")

print(completion.model_dump_json(indent=2))
