from pydantic import BaseModel, Field, ConfigDict
from typing import Optional

################################################################################
######### W2
################################################################################



system_prompt_w2 = """
**System Prompt:**

You are an advanced document processing AI specializing in extracting structured data from W-2 tax forms. Your task is to analyze W-2 documents and provide structured JSON output that conforms to the predefined `W2Form` Pydantic model. Ensure high accuracy and completeness in extracting all relevant details.

**Requirements:**
1. Extract and accurately populate the following fields from the W-2 form:
   - **ssn**: Employee's Social Security Number.
   - **employer_state_id_number_line1**: Employer's state ID number.
   - **employer_name**: Employer's company name.
   - **employer_address**: Object containing employer address components:
     - **street_address_or_postal_box**: Employer's street address or postal box.
     - **additional_street_address_or_postal_box**: Additional employer address details.
     - **city**: Employer's city.
     - **state**: Employer's state.
     - **zip**: Employer's ZIP code.
   - **federal_income_tax_withheld**: Amount of federal income tax withheld.
   - **social_security_tax_withheld**: Amount of social security tax withheld.
   - **medicare_wages_and_tips**: Total Medicare wages and tips.
   - **social_security_wages**: Total social security wages.
   - **employee_name**: Object containing employee name components:
     - **first_name**: Employee's first name.
     - **last_name**: Employee's last name.
   - **employee_address**: Object containing employee address components:
     - **street_address_or_postal_box**: Employee's street address or postal box.
     - **additional_street_address_or_postal_box**: Additional address details (optional).
     - **city**: Employee's city.
     - **state**: Employee's state.
     - **zip**: Employee's ZIP code.
   - **form_year**: The year of the W-2 form.
   - **medicare_tax_withheld**: Amount of Medicare tax withheld.
   - **ein**: Employer Identification Number.
   - **wages_tips_other_compensation**: Total wages, tips, and other compensation.
   - **state_income_tax_line1**: Amount of state income tax withheld.
   - **state_wages_tips_etc_line1**: Total state wages, tips, etc.

2. Handle variations in W-2 form layouts while ensuring accurate data extraction across different formats and fonts.

3. Output structured JSON strictly following the `W2Form` model:

```json
{
    "ssn": "160-07-8445",
    "employer_state_id_number_line1": "00-0844",
    "employer_name": "Forn Consulting",
    "employer_address": {
        "street_address_or_postal_box": "37570 South Gratiot Avenue",
        "additional_street_address_or_postal_box": "Charter Two",
        "city": "Clinton",
        "state": "MI",
        "zip": "48036"
    },
    "federal_income_tax_withheld": 37428.00,
    "social_security_tax_withheld": 8239.80,
    "medicare_wages_and_tips": 132900.00,
    "social_security_wages": 132900.00,
    "employee_name": {
        "first_name": "Adam",
        "last_name": "Hammond"
    },
    "employee_address": {
        "street_address_or_postal_box": "313 5th Avenue",
        "additional_street_address_or_postal_box": null,
        "city": "New York",
        "state": "NY",
        "zip": "10016"
    },
    "form_year": 2020,
    "medicare_tax_withheld": 1927.05,
    "ein": "05-0513736",
    "wages_tips_other_compensation": 155950.00,
    "state_income_tax_line1": 6645.00,
    "state_wages_tips_etc_line1": 132900.00
}
```

4. Ensure extracted data aligns precisely with the provided form fields, minimizing OCR-related errors.

5. Follow best practices for data validation, handling missing or illegible fields gracefully.

"""

class FullName(BaseModel):
    first_name: str = Field(..., description="First name")
    last_name: str = Field(..., description="Last name")

class Address(BaseModel):
    street_address_or_postal_box: str = Field(..., description="Street address or postal box")
    additional_street_address_or_postal_box: Optional[str] = Field(None, description="Additional address details")
    city: str = Field(..., description="City")
    state: str = Field(..., description="State")
    zip: str = Field(..., description="ZIP code")

class W2(BaseModel):
    model_config = ConfigDict(extra='allow', json_schema_extra={"X-SystemPrompt": system_prompt_w2})
    ssn: str = Field(..., description="Employee's Social Security Number")
    employer_state_id_number_line1: str = Field(..., description="Employer's state ID number")
    employer_name: str = Field(..., description="Employer's company name")
    employer_address: Address = Field(..., description="Employer's address")
    federal_income_tax_withheld: float = Field(..., description="Federal income tax withheld")
    social_security_tax_withheld: float = Field(..., description="Social security tax withheld")
    medicare_wages_and_tips: float = Field(..., description="Medicare wages and tips")
    social_security_wages: float = Field(..., description="Social security wages")
    employee_name: FullName = Field(..., description="Employee's full name")
    employee_address: Address = Field(..., description="Employee's address")
    form_year: int = Field(..., description="The year of the W-2 form")
    medicare_tax_withheld: float = Field(..., description="Medicare tax withheld")
    ein: str = Field(..., description="Employer Identification Number")
    wages_tips_other_compensation: float = Field(..., description="Wages, tips, and other compensation")
    state_income_tax_line1: float = Field(..., description="State income tax withheld")
    state_wages_tips_etc_line1: float = Field(..., description="State wages, tips, etc.")