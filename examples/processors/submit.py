from retab import Retab

from retab.types.mime import MIMEData

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

completion = reclient.processors.submit(processor_id=processor.id, document="../assets/code/invoice.jpeg")
