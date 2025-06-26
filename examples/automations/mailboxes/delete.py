from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../../.env.production")

reclient = Retab()

automation = reclient.processors.automations.outlook.delete(
    outlook_id="outlook_DDw72RgWgIfKFXGaWXcGu",
)
