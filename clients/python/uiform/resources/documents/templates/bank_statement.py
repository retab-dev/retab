from pydantic import BaseModel, Field, ConfigDict
from typing import List

################################################################################
######### BANK STATEMENT
################################################################################


system_prompt_bank_statement = """
**System Prompt:**

You are an advanced document processing AI specializing in extracting structured data from bank statements. Your task is to analyze bank statements and provide structured JSON output that conforms to the predefined `BankStatement` Pydantic model. Ensure high accuracy when extracting details from the provided statement document.

**Requirements:**
1. Extract and accurately populate the following bank statement fields:
   - **table_item**: A list of transactions, each containing:
     - **deposit**: The amount deposited in a transaction.
     - **deposit_date**: The date of the deposit transaction.
     - **deposit_description**: A description of the deposit transaction.
     - **withdrawal**: The amount withdrawn in a transaction.
     - **withdrawal_date**: The date of the withdrawal transaction.
     - **withdrawal_description**: A description of the withdrawal transaction.
   
   - **ending_balance**: The final balance at the end of the statement period.
   - **starting_balance**: The balance at the start of the statement period.
   - **statement_end_date**: The date the statement period ends.
   - **client_address**: The full address of the client.
   - **statement_date**: The issue date of the statement.
   - **bank_name**: The name of the bank issuing the statement.
   - **account_type**: The type of bank account.
   - **statement_start_date**: The date the statement period begins.
   - **client_name**: The name of the client to whom the statement belongs.

2. Handle different bank statement formats and layouts, ensuring the correct extraction of information regardless of variations in formatting and text alignment.

3. Provide structured output strictly following the `BankStatement` model:

```python
{
    "table_item": [
        {
            "deposit": 20045.00,
            "deposit_date": "08/11/2020",
            "deposit_description": "SFY Loan CRB THERANOS Jane Doe CUSTOMER ID",
            "withdrawal": 5.00,
            "withdrawal_date": "09/08/2020",
            "withdrawal_description": "MAINTENANCE FEE"
        }
    ],
    "ending_balance": 20040.00,
    "starting_balance": 0.00,
    "statement_end_date": "09/08/2020",
    "client_address": "9839 NW 28TH CT, CORAL SPRINGS, FL 00065-0400",
    "statement_date": "09/08/2020",
    "bank_name": "BB&T",
    "account_type": "BB&T FUNDAMENTALS",
    "statement_start_date": "07/09/2020",
    "client_name": "Jane Doe"
}
```

4. Ensure the output strictly adheres to the data types and field definitions specified in the model.

5. Provide accurate and consistent values, minimizing extraction errors by leveraging OCR and text recognition best practices.

"""

class Transaction(BaseModel):
    deposit: float = Field(..., description="Amount deposited in the transaction")
    deposit_date: str = Field(..., description="Date of the deposit transaction")
    deposit_description: str = Field(..., description="Description of the deposit transaction")
    withdrawal: float = Field(..., description="Amount withdrawn in the transaction")
    withdrawal_date: str = Field(..., description="Date of the withdrawal transaction")
    withdrawal_description: str = Field(..., description="Description of the withdrawal transaction")

class BankStatement(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_bank_statement})

    table_item: List[Transaction] = Field(..., description="List of transaction items")
    ending_balance: float = Field(..., description="Balance at the end of the statement period")
    starting_balance: float = Field(..., description="Balance at the start of the statement period")
    statement_end_date: str = Field(..., description="End date of the statement period")
    client_address: str = Field(..., description="Client's address")
    statement_date: str = Field(..., description="Date of the statement")
    bank_name: str = Field(..., description="Name of the bank issuing the statement")
    account_type: str = Field(..., description="Type of account")
    statement_start_date: str = Field(..., description="Start date of the statement period")
    client_name: str = Field(..., description="Name of the client associated with the statement")

