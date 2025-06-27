from retab import Retab

client = Retab()

mailbox = client.processors.automations.mailboxes.create(
    name="Invoice Mailbox",
    email="invoices@mailbox.retab.com",
    processor_id="proc_01G34H8J2K",  # The processor id you created in the previous step
    webhook_url="https://your-server.com/webhook",  # Replace with your actual webhook URL
)

# If you just want to send a test request to your webhook
log = client.processors.automations.tests.webhook(
    automation_id=mailbox.id,
)

# If you want to test the file processing logic:
log = client.processors.automations.tests.upload(automation_id=mailbox.id, document="your_invoice_email.eml")

# If you want to test a full email forwarding
log = client.processors.automations.mailboxes.tests.forward(email="invoices@mailbox.retab.com", document="your_invoice_email.eml")
