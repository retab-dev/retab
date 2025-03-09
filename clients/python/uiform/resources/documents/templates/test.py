from pydantic import BaseModel, Field, ConfigDict

################################################################################
######### PASSPORT
################################################################################

system_prompt_passport_json = """
**System Prompt:**

You are an advanced document processing AI specializing in extracting structured data from passports. Your task is to analyze passport documents and provide structured JSON output that conforms to the predefined passport schema. Ensure high accuracy and completeness in extracting all relevant details from the provided passport.

**Requirements:**

1. Extract and accurately populate all fields from the passport including:
   - Document type
   - Passport number
   - Country code
   - Nationality
   - Surname
   - Given names
   - Sex
   - Date of birth
   - Place of birth
   - Height
   - Eye color
   - Date of issue
   - Date of expiry
   - Issuing authority
   - MRZ data (machine-readable zone)
   - Signature

2. Handle different passport formats and layouts, ensuring accurate extraction across varying text alignments and fonts.

3. Ensure extracted data is validated against the provided schema, minimizing OCR-related errors.

4. Handle missing or illegible fields gracefully, ensuring partial but accurate data extraction where possible.
"""

class MRZ(BaseModel):
    line1: str = Field(..., description="First line of the MRZ")
    line2: str = Field(..., description="Second line of the MRZ")

class PassportJson(BaseModel):
    model_config = ConfigDict(extra='forbid', json_schema_extra={"X-SystemPrompt": system_prompt_passport_json})
    document_type: str = Field(..., description="Denotes the type of document, e.g., 'P' for passport")
    passport_number: str = Field(..., description="The unique passport number")
    country_code: str = Field(..., description="The ISO country code of the issuing country")
    nationality: str = Field(..., description="Holder's nationality")
    surname: str = Field(..., description="Holder's family name")
    given_names: str = Field(..., description="Holder's given name(s)")
    sex: str = Field(..., description="Holder's gender (e.g., M/F/X)")
    date_of_birth: str = Field(..., description="Holder's date of birth")
    place_of_birth: str = Field(..., description="Holder's place of birth")
    height: str = Field(..., description="Holder's height")
    eye_color: str = Field(..., description="Holder's eye color")
    date_of_issue: str = Field(..., description="Passport issue date")
    date_of_expiry: str = Field(..., description="Expiry date of the passport")
    issuing_authority: str = Field(..., description="Name of the authority that issued the passport")
    mrz: MRZ = Field(..., description="Machine-readable zone data")
    signature: str = Field(..., description="Holder's signature or reference")
