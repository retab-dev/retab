from pydantic import BaseModel, Field, ConfigDict
from typing import List

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
