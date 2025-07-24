import os
import argparse
from dotenv import load_dotenv
from retab import Retab

# File Path
file_path = os.path.dirname(os.path.abspath(__file__))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Create an Outlook automation")
    parser.add_argument("--env", required=True, choices=["dev", "staging", "prod"],
                       default="local",
                       help="Source environment")
    parser.add_argument("--processor-id", required=True,
                       help="ID of processor to create automation for")
    parser.add_argument("--webhook-url", required=False,
                       help="Webhook URL to send automation results to",
                       default="https://your-server.com/webhook")
    
    args = parser.parse_args()
    
    assert load_dotenv(f"{file_path}/../../../.env.{args.env}")

    reclient = Retab()

    automation = reclient.processors.automations.outlook.create(
        name="Outlook Automation",
        processor_id=args.processor_id,
        webhook_url=args.webhook_url,
    )

    print(automation.model_dump_json(indent=2))
