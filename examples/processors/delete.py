from curses import reset_shell_mode
from dotenv import load_dotenv
from retab import Retab

assert load_dotenv("../../.env.production")

reclient = Retab()

reclient.processors.delete(
    processor_id="proc_F0FE8DFqyouQdZXDTWRg0",
)
