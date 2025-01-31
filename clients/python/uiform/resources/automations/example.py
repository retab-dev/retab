from uiform import UiForm

uiclient = UiForm()

extraction_link = uiclient.automations.extraction_links.create(
    name="Invoices",
    json_schema="json_schema.json",
    webhook_url="https://your_server.com/invoices/webhook"
)
