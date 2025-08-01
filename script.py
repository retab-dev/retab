from retab import Retab
from dotenv import load_dotenv

load_dotenv(".env")

client = Retab()

extraction = client.documents.extract(json_schema="assets/code/invoice_schema.json", document="assets/code/invoice.jpeg", modality="native", model="gpt-4.1-nano", temperature=0)

print("Extraction: ", extraction.model_dump_json(indent=2))
