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
