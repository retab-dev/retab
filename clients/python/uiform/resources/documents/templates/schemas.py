from pydantic import BaseModel, Field, HttpUrl
from typing import List
from pydantic import ConfigDict
from pydantic import BaseModel, Field
from typing import List, Optional
from pydantic_core import Url
from pydantic import field_serializer

from pydantic import BaseModel, Field, HttpUrl
from typing import List
from pydantic import ConfigDict


################################################################################
######### BANK STATEMENT
################################################################################


system_prompt_bank_statement = """
**System Prompt:**

You are an advanced document processing AI specializing in extracting structured data from bank statements. Your task is to analyze bank statements and provide structured JSON output that conforms to the predefined `BankStatement` Pydantic model. Ensure high accuracy when extracting details from the provided statement document.

**Requirements:**
1. Extract and accurately populate the following bank statement fields:
   - **table_item**: A list of transactions, each containing:
     - **transaction_deposit**: The amount deposited in a transaction.
     - **transaction_deposit_date**: The date of the deposit transaction.
     - **transaction_deposit_description**: A description of the deposit transaction.
     - **transaction_withdrawal**: The amount withdrawn in a transaction.
     - **transaction_withdrawal_date**: The date of the withdrawal transaction.
     - **transaction_withdrawal_description**: A description of the withdrawal transaction.
   
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
            "transaction_deposit": 20045.00,
            "transaction_deposit_date": "08/11/2020",
            "transaction_deposit_description": "SFY Loan CRB THERANOS Jane Doe CUSTOMER ID",
            "transaction_withdrawal": 5.00,
            "transaction_withdrawal_date": "09/08/2020",
            "transaction_withdrawal_description": "MAINTENANCE FEE"
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
    transaction_deposit: float = Field(..., description="Amount deposited in the transaction")
    transaction_deposit_date: str = Field(..., description="Date of the deposit transaction")
    transaction_deposit_description: str = Field(..., description="Description of the deposit transaction")
    transaction_withdrawal: float = Field(..., description="Amount withdrawn in the transaction")
    transaction_withdrawal_date: str = Field(..., description="Date of the withdrawal transaction")
    transaction_withdrawal_description: str = Field(..., description="Description of the withdrawal transaction")

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




################################################################################
######### CONTRACT
################################################################################

system_prompt_contract = """
You are an advanced document processing AI specializing in extracting structured data from legal contracts. Your task is to analyze contract documents and provide structured JSON output that conforms to the predefined `Contract` Pydantic model. Ensure accuracy in capturing all relevant details from the contract document.

**Requirements:**
1. Extract and accurately populate the following contract fields:
   - **initial_term**: The initial duration of the contract.
   - **governing_law**: The legal jurisdiction governing the contract.
   - **agreement_date**: The date when the agreement was signed.
   - **renewal_term**: The terms under which the contract may be renewed.
   - **effective_date**: The date when the contract becomes effective.
   - **notice_to_terminate_renewal**: The required notice period to terminate or renew the contract.
   - **parties**: A list of parties involved in the contract.
   - **document_name**: The name of the contract document.

2. Handle different contract formats and ensure robustness in field extraction, considering varying structures, formatting styles, and legal terminologies.

3. Output must be formatted according to the `Contract` model:

```python
{
    "initial_term": "12 months",
    "governing_law": "State of California",
    "agreement_date": "2023-06-15",
    "renewal_term": "Automatic renewal for 6 months unless terminated",
    "effective_date": "2023-07-01",
    "notice_to_terminate_renewal": "30 days prior to renewal",
    "parties": ["Company A", "Company B"],
    "document_name": "Service Agreement 2023"
}
```

4. Ensure the output strictly adheres to the data types and field definitions specified in the model.

5. Provide accurate and consistent values, minimizing extraction errors by leveraging OCR and text recognition best practices.

"""

class Contract(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_contract})

    initial_term: str = Field(..., description="The initial duration of the contract.")
    governing_law: str = Field(..., description="The legal jurisdiction governing the contract.")
    agreement_date: str = Field(..., description="The date when the agreement was signed.")
    renewal_term: str = Field(..., description="The terms under which the contract may be renewed.")
    effective_date: str = Field(..., description="The date when the contract becomes effective.")
    notice_to_terminate_renewal: str = Field(..., description="The required notice period to terminate or renew the contract.")
    parties: List[str] = Field(..., description="The list of parties involved in the contract.")
    document_name: str = Field(..., description="The name of the contract document.")



################################################################################
######### DRIVER LICENSE
################################################################################



system_prompt_driver_license = """
You are an advanced document processing AI specializing in extracting structured data from driver's licenses. Your task is to analyze driver's license documents and provide structured JSON output that conforms to the predefined `DriverLicense` Pydantic model. Ensure high accuracy and completeness in extracting all relevant details from the provided document.

**Requirements:**

1. Extract and accurately populate the following fields from the driver's license:
   - **address**: The full address of the license holder.
   - **date_of_birth**: The birth date of the license holder.
   - **document_id**: The unique identifier of the driver's license.
   - **expiration_date**: The expiration date of the driver's license.
   - **family_name**: The last name (surname) of the license holder.
   - **given_names**: The first and middle names of the license holder.
   - **issue_date**: The date when the driver's license was issued.
   - **portrait**: A URL link to the portrait image of the license holder.

2. Handle different license formats and layouts, ensuring accurate extraction across varying text alignments and fonts.

3. Provide structured output strictly following the `DriverLicense` model:

```json
{
    "address": "175 Marvon Ave, Dallas, TX 75243",
    "date_of_birth": "06/27/1978",
    "document_id": "97551579",
    "expiration_date": "12/07/2041",
    "family_name": "Miller",
    "given_names": "Joseph John",
    "issue_date": "05/01/2031",
    "portrait": "https://example.com/portrait.jpg"
}
```

4. Ensure extracted data is validated against the provided schema, minimizing OCR-related errors.

5. Handle missing or illegible fields gracefully, ensuring partial but accurate data extraction where possible.
"""

class DriverLicense(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_driver_license})

    address: str = Field(..., description="The address of the license holder.")
    date_of_birth: str = Field(..., description="The date of birth of the license holder.")
    document_id: str = Field(..., description="The unique document ID of the driver's license.")
    expiration_date: str = Field(..., description="The expiration date of the driver's license.")
    family_name: str = Field(..., description="The last name of the license holder.")
    given_names: str = Field(..., description="The first and middle names of the license holder.")
    issue_date: str = Field(..., description="The issue date of the driver's license.")




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


################################################################################
######### IDENTITY PROOFING
################################################################################

system_prompt_identity_proofing = """
You are an expert document analysis AI specializing in identity proofing. Your task is to analyze identity documents and extract relevant fraud detection signals and evidence. 

**Your output should be structured according to the `IdentityProofingAnalysis` Pydantic model, with the following requirements:**

1. **Fraud Signals:** Identify potential fraud-related aspects of the document, including:
    - Whether the document is an identity document.
    - Suspicious words that may indicate fraud.
    - Any signs of image manipulation.
    - Whether the document appears duplicated online.
    - Detection of photocopies.

2. **Evidence:** Provide supporting details, such as:
    - A list of suspicious words found in the document.
    - Any inconclusive suspicious words requiring further review.
    - A URL to a thumbnail image of the document.
    - The hostname where the document was found online.

**Input Example:**

An image or scanned document of an identity card or passport.

**Expected Output:**

```python
IdentityProofing(
    fraud_signals=FraudSignals(
        is_identity_document=True,
        suspicious_words=["fake", "altered"],
        image_manipulation=True,
        online_duplicate=False,
        photocopy_detection=True
    ),
    evidence=Evidence(
        suspicious_words=["fake"],
        inconclusive_suspicious_word="altered",
        hostname="example.com"
    )
)
```

Ensure the output is structured, accurate, and aligns with the defined model while maintaining high reliability in fraud detection.

"""

class FraudSignals(BaseModel):
    is_identity_document: list[bool] = Field(default_factory=list, description="List of indicators if the document is identified as an identity document")
    suspicious_words: list[str] = Field(default_factory=list, description="List of words flagged as suspicious in the document")
    image_manipulation: list[bool] = Field(default_factory=list, description="List of flags indicating if the document image shows signs of manipulation")
    online_duplicate: list[bool] = Field(default_factory=list, description="List of flags indicating if the document appears duplicated online")
    photocopy_detection: list[bool] = Field(default_factory=list, description="List of flags indicating if the document is detected as a photocopy")

class Evidence(BaseModel):
    suspicious_words: list[str] = Field(default_factory=list, description="List of suspicious words found in the document")
    inconclusive_suspicious_word: list[Optional[str]] = Field(default_factory=list, description="List of words flagged as suspicious but inconclusive")
    hostname: list[Optional[str]] = Field(default_factory=list, description="List of hostnames where the document was found online")

class IdentityProofing(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_identity_proofing})
    
    fraud_signals: FraudSignals = Field(..., description="Fraud detection signals related to the document")
    evidence: Evidence = Field(..., description="Supporting evidence for the detected fraud signals")



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




################################################################################
######### PASSPORT
################################################################################



system_prompt_passport = """

**System Prompt:**

You are an advanced document processing AI specializing in extracting structured data from passports. Your task is to analyze passport documents and provide structured JSON output that conforms to the predefined `Passport` Pydantic model. Ensure high accuracy and completeness in extracting all relevant details from the provided passport.

**Requirements:**

1. Extract and accurately populate the following fields from the passport:
   - **date_of_birth**: The birth date of the passport holder.
   - **document_id**: The unique identifier of the passport.
   - **expiration_date**: The expiration date of the passport.
   - **family_name**: The last name (surname) of the passport holder.
   - **given_names**: The first and middle names of the passport holder.
   - **issue_date**: The date when the passport was issued.
   - **mrz_code**: The machine-readable zone (MRZ) code from the passport.
   - **portrait**: A URL link to the portrait image of the passport holder.

2. Handle different passport formats and layouts, ensuring accurate extraction across varying text alignments and fonts.

3. Provide structured output strictly following the `Passport` model:

```json
{
    "date_of_birth": "05 FEB 1965",
    "document_id": "E00007730",
    "expiration_date": "14 OCT 2030",
    "family_name": "TRAVELER",
    "given_names": "HAPPY",
    "issue_date": "15 OCT 2020",
    "mrz_code": "P<USATRAVELER<<HAPPY<<<<<<<<<<<<<<<<<<<\nE000077303USA6502056F3010149500101920<091824",
}
```

4. Ensure extracted data is validated against the provided schema, minimizing OCR-related errors.

5. Handle missing or illegible fields gracefully, ensuring partial but accurate data extraction where possible.

"""

class Passport(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_passport})
    date_of_birth: str = Field(..., description="The date of birth of the passport holder.")
    document_id: str = Field(..., description="The unique document ID of the passport.")
    expiration_date: str = Field(..., description="The expiration date of the passport.")
    family_name: str = Field(..., description="The last name of the passport holder.")
    given_names: str = Field(..., description="The first and middle names of the passport holder.")
    issue_date: str = Field(..., description="The issue date of the passport.")
    mrz_code: str = Field(..., description="The machine-readable zone (MRZ) code from the passport.")


################################################################################
######### PAY SLIP
################################################################################

system_prompt_pay_slip = """
You are an advanced document processing AI specializing in extracting structured data from pay slips. Your task is to analyze pay slip documents and provide structured JSON output that conforms to the predefined `PaySlip` Pydantic model. Ensure high accuracy and completeness in extracting all relevant details from the provided pay slip.

**Requirements:**

1. Extract and accurately populate the following fields from the pay slip:
   - **employer_address**: Full address of the employer.
   - **gross_earnings**: Total gross earnings for the period.
   - **deductions**: A list of deductions, each containing:
     - **deduction_this_period**: Amount deducted in the current period.
     - **deduction_type**: Type of deduction.
     - **deduction_ytd**: Year-to-date deduction amount.
   - **federal_marital_status**: Employee's federal marital status.
   - **employee_name**: Full name of the employee.
   - **gross_earnings_ytd**: Year-to-date gross earnings.
   - **end_date**: End date of the pay period.
   - **employer_name**: Name of the employer.
   - **pay_date**: Date when payment was issued.
   - **state_allowance**: Number of state allowances claimed.
   - **ssn**: Employee's Social Security Number.
   - **employee_address**: Full address of the employee.
   - **taxes**: A list of tax details, each containing:
     - **tax_this_period**: Tax amount deducted in the current period.
     - **tax_type**: Type of tax.
     - **tax_ytd**: Year-to-date tax amount.
   - **federal_allowance**: Number of federal allowances claimed.
   - **earnings**: A list of earning items, each containing:
     - **earning_hours**: Hours worked for this earning item.
     - **earning_rate**: Hourly rate for this earning item.
     - **earning_this_period**: Amount earned in the current period.
     - **earning_type**: Type of earning.
     - **earning_ytd**: Year-to-date earning amount.
   - **net_pay**: Net pay after deductions.

2. Handle variations in pay slip layouts and ensure accurate extraction across different formats and styles.

3. Provide structured output strictly following the `PaySlip` model:

```json
{
    "employer_address": "70 Maple tree Business City, Singapore 117371",
    "gross_earnings": 1085.50,
    "deductions": [
        {"deduction_this_period": 86.84, "deduction_type": "Federal Income Tax", "deduction_ytd": 260.52},
        {"deduction_this_period": 65.13, "deduction_type": "Social Security Tax", "deduction_ytd": 195.39}
    ],
    "federal_marital_status": "Single",
    "employee_name": "Dave Petty",
    "gross_earnings_ytd": 3265.50,
    "end_date": "12/31/2014",
    "employer_name": "Google Singapore",
    "pay_date": "01/07/2015",
    "state_allowance": 3,
    "ssn": "015-04-3122",
    "employee_address": "208 E. Main Street, Bozeman MT 59715",
    "taxes": [
        {"tax_this_period": 21.71, "tax_type": "Medicare Tax", "tax_ytd": 65.13}
    ],
    "federal_allowance": 4,
    "earnings": [
        {"earning_hours": 37.00, "earning_rate": 20.00, "earning_this_period": 740.00, "earning_type": "Regular", "earning_ytd": 2220.00}
    ],
    "net_pay": 742.20
}
```

4. Ensure extracted data is validated against the provided schema, minimizing OCR-related errors.

5. Handle missing or illegible fields gracefully, ensuring partial but accurate data extraction where possible.
"""

class DeductionItem(BaseModel):
    deduction_this_period: float = Field(..., description="Amount deducted this period")
    deduction_type: str = Field(..., description="Type of deduction")
    deduction_ytd: float = Field(..., description="Year-to-date deduction amount")

class TaxItem(BaseModel):
    tax_this_period: float = Field(..., description="Tax amount deducted this period")
    tax_type: str = Field(..., description="Type of tax")
    tax_ytd: float = Field(..., description="Year-to-date tax amount")

class EarningItem(BaseModel):
    earning_hours: float = Field(..., description="Hours worked for this earning item")
    earning_rate: float = Field(..., description="Hourly rate for this earning item")
    earning_this_period: float = Field(..., description="Amount earned this period")
    earning_type: str = Field(..., description="Type of earning")
    earning_ytd: float = Field(..., description="Year-to-date earning amount")

class PaySlip(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_pay_slip})
    employer_address: str = Field(..., description="Address of the employer")
    gross_earnings: float = Field(..., description="Gross earnings for the period")
    deductions: List[DeductionItem] = Field(..., description="List of deductions applied")
    federal_marital_status: str = Field(..., description="Federal marital status of the employee")
    employee_name: str = Field(..., description="Name of the employee")
    gross_earnings_ytd: float = Field(..., description="Year-to-date gross earnings")
    end_date: str = Field(..., description="End date of the pay period")
    employer_name: str = Field(..., description="Name of the employer")
    pay_date: str = Field(..., description="Date the payment was issued")
    state_allowance: int = Field(..., description="State allowances claimed")
    ssn: str = Field(..., description="Social Security Number of the employee")
    employee_address: str = Field(..., description="Address of the employee")
    taxes: List[TaxItem] = Field(..., description="List of taxes applied")
    federal_allowance: int = Field(..., description="Federal allowances claimed")
    earnings: List[EarningItem] = Field(..., description="List of earnings")
    net_pay: float = Field(..., description="Net pay after deductions")


################################################################################
######### W2
################################################################################



system_prompt_w2 = """
**System Prompt:**

You are an advanced document processing AI specializing in extracting structured data from W-2 tax forms. Your task is to analyze W-2 documents and provide structured JSON output that conforms to the predefined `W2Form` Pydantic model. Ensure high accuracy and completeness in extracting all relevant details.

**Requirements:**

1. Extract and accurately populate the following fields from the W-2 form:
   - **SSN**: Employee's Social Security Number.
   - **EmployerStateIdNumber_Line1**: Employer's state ID number.
   - **EmployeeName_FirstName**: Employee's first name.
   - **EmployerAddress_StreetAddressOrPostalBox**: Employer's street address or postal box.
   - **FederalIncomeTaxWithheld**: Amount of federal income tax withheld.
   - **EmployeeFirstNameAndInitial**: Employee's first name and initial.
   - **EmployerNameAndAddress**: Full employer's name and address.
   - **SocialSecurityTaxWithheld**: Amount of social security tax withheld.
   - **MedicareWagesAndTips**: Total Medicare wages and tips.
   - **EmployeeName_LastName**: Employee's last name.
   - **State_Line1**: State abbreviation.
   - **SocialSecurityWages**: Total social security wages.
   - **EmployeeAddress_State**: Employee's state.
   - **EmployeeAddress_StreetAddressOrPostalBox**: Employee's street address or postal box.
   - **EmployerAddress_City**: Employer's city.
   - **EmployeeAddress**: Full employee's address.
   - **FormYear**: The year of the W-2 form.
   - **EmployeeLastName**: Employee's last name.
   - **EmployerName**: Employer's name.
   - **MedicareTaxWithheld**: Amount of Medicare tax withheld.
   - **EmployerAddress_State**: Employer's state.
   - **EmployeeName**: Full employee's name.
   - **EIN**: Employer Identification Number.
   - **WagesTipsOtherCompensation**: Total wages, tips, and other compensation.
   - **EmployerAddress_AdditionalStreetAddressOrPostalBox**: Additional employer address details.
   - **EmployeeAddress_City**: Employee's city.
   - **EmployeeAddress_Zip**: Employee's ZIP code.
   - **EmployerAddress_Zip**: Employer's ZIP code.
   - **StateIncomeTax_Line1**: Amount of state income tax withheld.
   - **StateWagesTipsEtc_Line1**: Total state wages, tips, etc.

2. Handle variations in W-2 form layouts while ensuring accurate data extraction across different formats and fonts.

3. Output structured JSON strictly following the `W2Form` model:

```json
{
    "SSN": "160-07-8445",
    "EmployerStateIdNumber_Line1": "00-0844",
    "EmployeeName_FirstName": "Adam",
    "EmployerAddress_StreetAddressOrPostalBox": "37570 South Gratiot Avenue",
    "FederalIncomeTaxWithheld": 37428.00,
    "EmployeeFirstNameAndInitial": "Adam",
    "EmployerNameAndAddress": "Forn Consulting, 37570 South Gratiot Avenue, Clinton MI 48036",
    "SocialSecurityTaxWithheld": 8239.80,
    "MedicareWagesAndTips": 132900.00,
    "EmployeeName_LastName": "Hammond",
    "State_Line1": "MI",
    "SocialSecurityWages": 132900.00,
    "EmployeeAddress_State": "NY",
    "EmployeeAddress_StreetAddressOrPostalBox": "313 5th Avenue",
    "EmployerAddress_City": "Clinton",
    "EmployeeAddress": "313 5th Avenue, New York NY 10016",
    "FormYear": 2020,
    "EmployeeLastName": "Hammond",
    "EmployerName": "Forn Consulting",
    "MedicareTaxWithheld": 1927.05,
    "EmployerAddress_State": "MI",
    "EmployeeName": "Adam Hammond",
    "EIN": "05-0513736",
    "WagesTipsOtherCompensation": 155950.00,
    "EmployerAddress_AdditionalStreetAddressOrPostalBox": "Charter Two",
    "EmployeeAddress_City": "New York",
    "EmployeeAddress_Zip": "10016",
    "EmployerAddress_Zip": "48036",
    "StateIncomeTax_Line1": 6645.00,
    "StateWagesTipsEtc_Line1": 132900.00
}
```

4. Ensure extracted data aligns precisely with the provided form fields, minimizing OCR-related errors.

5. Follow best practices for data validation, handling missing or illegible fields gracefully.

"""

class W2(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_w2})
    SSN: str = Field(..., description="Employee's Social Security Number")
    EmployerStateIdNumber_Line1: str = Field(..., description="Employer's state ID number")
    EmployeeName_FirstName: str = Field(..., description="Employee's first name")
    EmployerAddress_StreetAddressOrPostalBox: str = Field(..., description="Employer's street address or postal box")
    FederalIncomeTaxWithheld: float = Field(..., description="Federal income tax withheld")
    EmployeeFirstNameAndInitial: str = Field(..., description="Employee's first name and initial")
    EmployerNameAndAddress: str = Field(..., description="Employer's full name and address")
    SocialSecurityTaxWithheld: float = Field(..., description="Social security tax withheld")
    MedicareWagesAndTips: float = Field(..., description="Medicare wages and tips")
    EmployeeName_LastName: str = Field(..., description="Employee's last name")
    State_Line1: str = Field(..., description="State abbreviation")
    SocialSecurityWages: float = Field(..., description="Social security wages")
    EmployeeAddress_State: str = Field(..., description="Employee's state")
    EmployeeAddress_StreetAddressOrPostalBox: str = Field(..., description="Employee's street address or postal box")
    EmployerAddress_City: str = Field(..., description="Employer's city")
    EmployeeAddress: str = Field(..., description="Employee's full address")
    FormYear: int = Field(..., description="The year of the W-2 form")
    EmployeeLastName: str = Field(..., description="Employee's last name")
    EmployerName: str = Field(..., description="Employer's name")
    MedicareTaxWithheld: float = Field(..., description="Medicare tax withheld")
    EmployerAddress_State: str = Field(..., description="Employer's state")
    EmployeeName: str = Field(..., description="Employee's full name")
    EIN: str = Field(..., description="Employer Identification Number")
    WagesTipsOtherCompensation: float = Field(..., description="Wages, tips, and other compensation")
    EmployerAddress_AdditionalStreetAddressOrPostalBox: str = Field(..., description="Employer's additional address details")
    EmployeeAddress_City: str = Field(..., description="Employee's city")
    EmployeeAddress_Zip: str = Field(..., description="Employee's ZIP code")
    EmployerAddress_Zip: str = Field(..., description="Employer's ZIP code")
    StateIncomeTax_Line1: float = Field(..., description="State income tax")
    StateWagesTipsEtc_Line1: float = Field(..., description="State wages, tips, etc.")
