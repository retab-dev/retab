from pydantic import BaseModel, Field
from typing import Optional


system_prompt_credit_card = """Use the provided credit card schema to extract corresponding fields from photos of credit cards. Make sure to capture the card number, name, expiration date, CVV, network, subtype, and issuer information exactly as shown on the card. For expiration dates, adhere to the printed format (MM/YY or MM/YYYY). Ensure the data is faithfully transcribed without additional spaces or punctuation."""

class CreditCard(BaseModel):
    card_number: str = Field(
        description="The card number exactly as displayed on the front of the card",
        json_schema_extra={"X-FieldPrompt": "Locate the credit card number (usually embossed or printed on the front). Extract all digits in the correct sequence, ignoring any spaces or dashes."}
    )
    card_holder_name: str = Field(
        description="The name of the cardholder as it appears on the card",
        json_schema_extra={"X-FieldPrompt": "Identify and extract the cardholder name as it appears on the card, usually in uppercase letters on the front."}
    )
    network: str = Field(
        description="The card's payment network (e.g., Visa, MasterCard, American Express)",
        json_schema_extra={"X-FieldPrompt": "Determine the card's payment network (e.g., Visa, MasterCard, Amex) based on any logos or known BIN ranges."}
    )
    subtype: str = Field(
        description="The card's subtype or level (e.g., Platinum, Gold, etc.)",
        json_schema_extra={"X-FieldPrompt": "If visible, extract the card's subtype (such as Platinum, Gold), typically indicated near the network branding."}
    )
    expiration_date: str = Field(
        description="The expiration date printed on the card, in the format MM/YY or MM/YYYY",
        json_schema_extra={"X-FieldPrompt": "Locate the expiration date on the card, usually in MM/YY or MM/YYYY format. Capture the date exactly as shown on the card."}
    )
    cvv: str = Field(
        description="The card's security code (CVV/CVC)",
        json_schema_extra={"X-FieldPrompt": "If available, extract the three- or four-digit CVV/CVC code, typically found on the back of the card."}
    )
    issuer: str = Field(
        description="The financial institution or organization that issued the card",
        json_schema_extra={"X-FieldPrompt": "Identify the issuer's name or logo (e.g., the bank or financial institution) and extract it as presented."}
    )

    class Config:
        title = "Credit_Card_Schema"
        json_schema_extra = {
            "X-SystemPrompt": system_prompt_credit_card,
            "X-ReasoningPrompt": "Before extracting the structured fields, use this space to outline your plan. Resolve uncertainties such as missing or obscured details, confirm the card's network and subtype from any logos or text, and ensure the correct length and grouping for the card number. Decide how to handle potential inconsistencies (e.g., unusual date formats or incomplete names) and confirm that each extracted value corresponds accurately to the area indicated on the card image."
        } 