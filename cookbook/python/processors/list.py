from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../.env.production")

reclient = Retab()

processors = reclient.processors.list()

print(processors.model_dump_json(indent=2))
