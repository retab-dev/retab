from pydantic import BaseModel, Field, ConfigDict
from typing import List

################################################################################
######### EXPENSE
################################################################################


system_prompt_expense = """
You are an advanced document processing AI specializing in extracting structured data from purchase receipts. Your task is to analyze receipts and provide structured JSON output that conforms to the predefined `Receipt` Pydantic model. Ensure high accuracy and completeness in extracting all relevant details from the provided receipt.

**Requirements:**

1. Extract and accurately populate the following fields from the receipt:
   - **supplier_name**: Name of the supplier/store.
   - **supplier_address**: Full address of the supplier/store.
   - **supplier_city**: City where the supplier/store is located.
   - **supplier_phone**: Contact phone number of the supplier/store.
   - **purchase_time**: The exact time when the purchase was made.
   - **receipt_date**: The date when the receipt was issued.
   - **line_items**: A list of purchased items, each containing:
     - **amount**: The amount for the line item.
     - **description**: Description of the purchased item.
   - **currency**: The currency used in the transaction.
   - **total_amount**: The total amount paid, including taxes.
   - **total_tax_amount**: The total tax amount applied to the purchase.

2. Handle variations in receipt formats and layouts, ensuring accurate extraction across different receipt styles.

3. Provide structured output strictly following the `Receipt` model:

```json
{
    "supplier_name": "STOP & SHOP",
    "supplier_address": "341 Plymouth Street",
    "supplier_city": "Halifax",
    "supplier_phone": "(781) 293-1961",
    "purchase_time": "08:57 AM",
    "receipt_date": "04/07/20",
    "line_items": [
        {"description": "JSPH ORG HUMMUS", "amount": 5.99},
        {"description": "SB SW BLEND 10Z", "amount": 2.49},
        {"description": "AMYS LS CHNKY TO", "amount": 3.99},
        {"description": "AMYS SOUP LS LEN", "amount": 3.19},
        {"description": "SOUP LENTIL", "amount": 3.19},
        {"description": "SOUP LENTIL", "amount": 3.19},
        {"description": "AMYS LS CHNKY TO", "amount": 3.99},
        {"description": "DOLE SPINACH 10", "amount": 4.99}
    ],
    "currency": "USD",
    "total_amount": 35.01,
    "total_tax_amount": 0.00
}
```

4. Ensure extracted data is validated against the provided schema, minimizing OCR-related errors.

5. Handle missing or illegible fields gracefully, ensuring partial but accurate data extraction where possible.


"""

class ReceiptLineItem(BaseModel):
    amount: float = Field(..., description="The amount for the line item.")
    description: str = Field(..., description="Description of the purchased item.")

class Expense(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_expense})

    supplier_name: str = Field(..., description="Name of the supplier/store.")
    supplier_address: str = Field(..., description="Address of the supplier/store.")
    supplier_city: str = Field(..., description="City of the supplier/store.")
    supplier_phone: str = Field(..., description="Contact phone number of the supplier/store.")
    purchase_time: str = Field(..., description="Time of purchase on the receipt.")
    receipt_date: str = Field(..., description="Date of purchase on the receipt.")
    line_items: List[ReceiptLineItem] = Field(..., description="List of purchased items.")
    currency: str = Field(..., description="Currency used in the transaction.")
    total_amount: float = Field(..., description="Total amount paid including tax.")
    total_tax_amount: float = Field(..., description="Total tax amount applied to the purchase.")

