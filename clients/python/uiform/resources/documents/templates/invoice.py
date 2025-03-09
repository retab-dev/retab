from pydantic import BaseModel, Field, ConfigDict
from typing import List
from pydantic import HttpUrl
from pydantic import field_serializer

################################################################################
######### INVOICE
################################################################################

system_prompt_invoice = """
You are an advanced document processing AI specializing in extracting structured data from invoices. Your task is to analyze invoices and provide structured JSON output that conforms to the predefined `Invoice` Pydantic model. Ensure accuracy in capturing all relevant details from the invoice image or document.

**Requirements:**
1. Extract and accurately populate the following invoice fields:
   - Supplier information (name, phone, email, address, website)
   - Receiver details (name, phone, email, address)
   - Invoice details (date, ID, currency)
   - Financial information (total amount, tax amount, net amount, amount due, amount paid since last invoice)
   - Shipping details (ship-to name, address)
   - Line items, each containing:
     - Description
     - Quantity
     - Unit price
     - Amount

2. Handle different invoice formats and ensure robustness in field extraction, considering varying layouts and text alignments.

3. Output must be formatted according to the `Invoice` model:

```python
{
    "supplier_phone": "123-456-7890",
    "invoice_date": "2024-01-22",
    "amount_due": 68.01,
    "supplier_email": "sales@amnoshsuppliers.com",
    "receiver_phone": "321-321-1234",
    "total_tax_amount": 225.87,
    "invoice_id": "1437",
    "supplier_name": "AMNOSH SUPPLIERS",
    "total_amount": 12113.67,
    "receiver_name": "Johnson Carrie",
    "supplier_website": "http://www.amnoshsuppliers.com",
    "supplier_address": "9291 Proin Road, Lake Charles, ME-11292",
    "currency": "USD",
    "ship_to_name": "Johnny Patel",
    "ship_to_address": "45 Lightning Road, Arizona, AZ 88776",
    "line_items": [
        {"description": "Drag Series Transmission Build - A WD DSM", "quantity": 3, "unit_price": 1129.03, "amount": 3387.09},
        {"description": "Drive Shaft Automatic Right", "quantity": 2, "unit_price": 243.01, "amount": 486.02},
        {"description": "MIZOL 20W40 Engine Oil", "quantity": 4, "unit_price": 342.00, "amount": 1368.00},
        {"description": "Spirax W2 ATF", "quantity": 3, "unit_price": 54.50, "amount": 163.50},
        {"description": "Hydraulic Press-25 Tons", "quantity": 1, "unit_price": 6391.85, "amount": 6391.85},
        {"description": "Optional: Slotter Machine", "quantity": 2, "unit_price": 45.67, "amount": 91.34}
    ],
    "receiver_email": "proprietor@abcxyz.com",
    "receiver_address": "45 Lightning Road, Arizona, AZ 88776",
    "amount_paid_since_last_invoice": 12045.66,
    "net_amount": 68.01
}
```

4. Ensure the output strictly adheres to the data types and field definitions specified in the model.

5. Provide accurate and consistent values, minimizing extraction errors by leveraging OCR and text recognition best practices.
"""

class InvoiceLineItem(BaseModel):
    amount: float = Field(..., description="Total price for this line item")
    description: str = Field(..., description="Description of the item")
    quantity: int = Field(..., description="Quantity of the item purchased")
    unit_price: float = Field(..., description="Price per unit of the item")

class Invoice(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_invoice})

    supplier_phone: str = Field(..., description="Phone number of the supplier")
    invoice_date: str = Field(..., description="Date of the invoice")
    amount_due: float = Field(..., description="Total amount due after payments")
    supplier_email: str = Field(..., description="Email address of the supplier")
    receiver_phone: str = Field(..., description="Phone number of the receiver")
    total_tax_amount: float = Field(..., description="Total tax amount applied")
    invoice_id: str = Field(..., description="Unique identifier for the invoice")
    supplier_name: str = Field(..., description="Name of the supplier")
    total_amount: float = Field(..., description="Total invoice amount including tax")
    receiver_name: str = Field(..., description="Name of the invoice receiver")
    supplier_website: HttpUrl = Field(..., description="Website URL of the supplier")
    supplier_address: str = Field(..., description="Address of the supplier")
    currency: str = Field(..., description="Currency of the invoice amount")
    ship_to_name: str = Field(..., description="Name of the person the invoice is shipped to")
    ship_to_address: str = Field(..., description="Shipping address")
    line_items: List[InvoiceLineItem] = Field(..., description="List of line items in the invoice")
    receiver_email: str = Field(..., description="Email address of the receiver")
    receiver_address: str = Field(..., description="Address of the invoice receiver")
    amount_paid_since_last_invoice: float = Field(..., description="Amount paid since the last invoice")
    net_amount: float = Field(..., description="Net amount to be paid after deductions")

    @field_serializer('supplier_website')
    def url2str(self, val: HttpUrl) -> str:
        return str(val)

