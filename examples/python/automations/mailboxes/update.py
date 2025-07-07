from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.mailboxes.update(
    mailbox_id="mb_FRf6FX5fYkenZ_JJlL5GD",
    name="Mailbox Automation Updated",
)

print(automation.model_dump_json(indent=2))
