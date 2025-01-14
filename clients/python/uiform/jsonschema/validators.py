from typing import Optional, Any, Union
import stdnum.eu.vat  # type: ignore
import phonenumbers
import datetime
from email_validator import validate_email
import pycountry
import re

# **** Validation Functions ****

# 1) Special Objects

def validate_currency(currency_code: Any) -> Optional[str]:
    """
    Return the valid currency code (ISO 4217) or None if invalid.
    """
    if currency_code is None:
        return None
    currency_code = str(currency_code).strip()  # convert to str and trim
    if not currency_code:
        return None
    try:
        if pycountry.currencies.lookup(currency_code):
            return currency_code
    except LookupError:
        pass
    return None


def validate_country_code(v: Any) -> Optional[str]:
    """
    Return the valid country code (ISO 3166) or None if invalid.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        if pycountry.countries.lookup(v_str):
            return v_str
    except LookupError:
        pass
    return None


def validate_email_regex(v: Any) -> Optional[str]:
    """
    Return the string if it matches a basic email pattern, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    pattern = r'^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$'
    if re.match(pattern, v_str):
        return v_str.lower()
    return None


def validate_vat_number(v: Any) -> Optional[str]:
    """
    Return the VAT number if valid (EU format) else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        if stdnum.eu.vat.is_valid(v_str):
            return stdnum.eu.vat.validate(v_str)
    except:
        pass
    return None


def validate_phone_number(v: Any) -> Optional[str]:
    """
    Return E.164 phone number format if valid, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        phone_number = phonenumbers.parse(v_str, "FR")  # Default region: FR
        if phonenumbers.is_valid_number(phone_number):
            return phonenumbers.format_number(phone_number, phonenumbers.PhoneNumberFormat.E164)
    except phonenumbers.NumberParseException:
        pass
    return None


def validate_email_address(v: Any) -> Optional[str]:
    """
    Return the normalized email address if valid, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        return validate_email(v_str).normalized
    except:
        return None


def validate_frenchpostcode(v: Any) -> Optional[str]:
    """
    Return a 5-digit postcode if valid, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    # Zero-pad to 5 digits
    try:
        v_str = v_str.zfill(5)
        # Optionally check numeric
        if not v_str.isdigit():
            return None
        return v_str
    except:
        return None


def validate_packing_type(v: Any) -> Optional[str]:
    """
    Return the packing type if in the known set, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip().lower()
    # We’ll store the valid set in lower for easy comparison
    valid_packing_types = {'box', 'pallet', 'container', 'bag', 'drum', 'other'}
    if v_str in valid_packing_types:
        return v_str
    return None


def validate_un_code(v: Any) -> Optional[int]:
    """
    Return an integer UN code in range [0..3481], else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        val = int(float(v_str))  # handle numeric strings
        if 0 <= val <= 3481:
            return val
    except:
        pass
    return None


def validate_adr_tunnel_code(v: Any) -> Optional[str]:
    """
    Return a valid ADR tunnel code from a known set, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip().upper()  # unify for set comparison
    valid_codes = {
        'B', 'B1000C', 'B/D', 'B/E', 'C', 'C5000D', 'C/D', 'C/E',
        'D', 'D/E', 'E', '-'
    }
    return v_str if v_str in valid_codes else None


def validate_un_packing_group(v: Any) -> Optional[str]:
    """
    Return a valid UN packing group (I, II, or III), else None.
    """
    if v is None:
        return None
    v_str = str(v).strip().upper()
    valid_groups = {'I', 'II', 'III'}
    return v_str if v_str in valid_groups else None


# 2) General Objects

def validate_integer(v: Any) -> Optional[int]:
    """
    Return an integer if parseable, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        return int(float(v_str))
    except:
        return None


def validate_float(v: Any) -> Optional[float]:
    """
    Return a float if parseable, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if not v_str:
        return None
    try:
        return float(v_str)
    except:
        return None


def validate_date(v: Union[str, datetime.date, None]) -> Optional[str]:
    """
    Return date in ISO format (YYYY-MM-DD) if valid, else None.
    """
    if v is None:
        return None

    # If it's already a date object
    if isinstance(v, datetime.date):
        return v.isoformat()

    # If it's a string
    v_str = str(v).strip()
    if not v_str:
        return None

    # Try ISO or a close variant
    try:
        return datetime.date.fromisoformat(v_str).isoformat()
    except ValueError:
        # Fallback to strptime
        try:
            return datetime.datetime.strptime(v_str, "%Y-%m-%d").date().isoformat()
        except ValueError:
            return None


def validate_time(v: Union[str, datetime.time, None]) -> Optional[str]:
    """
    Return time in ISO format (HH:MM[:SS]) if valid, else None.
    """
    if v is None:
        return None

    # If it's already a time object
    if isinstance(v, datetime.time):
        return v.isoformat()

    v_str = str(v).strip()
    if not v_str:
        return None

    # Try multiple formats
    time_formats = ["%H:%M:%S", "%H:%M", "%I:%M %p", "%I:%M:%S %p"]
    for fmt in time_formats:
        try:
            parsed = datetime.datetime.strptime(v_str, fmt).time()
            return parsed.isoformat()
        except ValueError:
            continue
    return None


def validate_bool(v: Any) -> bool:
    """
    Convert to bool if matches known true/false strings or actual bool.
    Otherwise return False.
    """
    if v is None:
        return False

    if isinstance(v, bool):
        return v

    try:
        v_str = str(v).strip().lower()
        true_values = {"true", "t", "yes", "y", "1"}
        false_values = {"false", "f", "no", "n", "0"}
        if v_str in true_values:
            return True
        elif v_str in false_values:
            return False
    except:
        pass

    return False


def validate_strold(v: Any) -> Optional[str]:
    """
    Return a stripped string unless it's empty or a known 'null' placeholder, else None.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    # Treat these placeholders (and empty) as invalid
    if v_str.lower() in {'null', 'none', 'nan', ''}:
        return None
    return v_str

def validate_str(v: Any) -> Optional[str]:
    """
    Return a stripped string unless it's invalid (e.g., placeholders like 'null'), else None.
    Does NOT convert empty strings to None—leaves them as-is.
    """
    if v is None:
        return None
    v_str = str(v).strip()
    if v_str.lower() in {'null', 'none', 'nan'}:  # Only treat explicit placeholders as None
        return None
    return v_str  # Keep empty strings intact


def notnan(x: Any) -> bool:
    """
    Return False if x is None, 'null', 'nan', or x != x (NaN check).
    True otherwise.
    """
    if x is None:
        return False
    x_str = str(x).lower().strip()
    if x_str in {"null", "nan"}:
        return False
    # Check for actual float NaN (x != x)
    return not (x != x)