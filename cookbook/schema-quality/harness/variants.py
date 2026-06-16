"""Build the nullable and reasoning schema variants from the agent baseline.

The baseline (schemas/baseline.json) is what `retab schemas generate` produced
from one representative invoice: every field is `required` and no field may be
null. We derive two upgrades from it, changing one lever at a time so the
comparison is controlled:

  nullable   : the optional fields (the ones that are genuinely absent on some
               invoices) are retyped as ["<type>", "null"]. They stay in
               `required` — under strict structured output every property is
               emitted anyway, so optionality is expressed through the null
               type, not by dropping the key. Nothing else changes.

  reasoning  : the nullable schema PLUS an X-ReasoningPrompt on each optional
               field and an X-SystemPrompt on the whole schema.

What is a reasoning prompt?
  X-ReasoningPrompt is a Retab schema extension. The text is given to the model
  as a private, per-field instruction telling it HOW to decide the value before
  it answers (a "think first" hint). It is not stored in the output; it only
  steers extraction. Here we use it to say, in effect, "only return this field
  if it is actually printed; if it is absent, return null, and never copy a
  value from another field." X-SystemPrompt is the same idea at the whole-schema
  level.

Run:  python -m harness.variants     # writes schemas/nullable.json + reasoning.json
"""

from __future__ import annotations

import copy
import json
import os

HERE = os.path.dirname(os.path.abspath(__file__))
SCHEMAS_DIR = os.path.join(os.path.dirname(HERE), "schemas")

# Fields that are optional in this corpus (present on some invoices, absent on
# others). These are the only fields the levers touch.
OPTIONAL_FIELDS = [
    "buyer_name",
    "customer_code",
    "purchase_order_number",
    "discount_amount",
    "due_date",
    "tax_amount",
]

SYSTEM_PROMPT = (
    "You are extracting fields from an invoice. Invoices often omit optional "
    "fields. Only return a value that is actually printed on this invoice; when "
    "a field is absent, return null. Never copy a value from a different field "
    "to fill a gap. Dates are European day/month/year and must be returned in "
    "ISO 8601 (YYYY-MM-DD)."
)

REASONING_PROMPTS = {
    "buyer_name": (
        "Return the billed-to contact person's name. If the invoice is billed to "
        "a company with no named contact, return null. Do not use the company "
        "name as the contact name."
    ),
    "customer_code": (
        "Return the customer or account code only if one is explicitly printed "
        "(e.g. a 'Customer code' line). If none is shown, return null; do not "
        "guess or reuse the invoice number."
    ),
    "purchase_order_number": (
        "Return the purchase order number only if a PO is explicitly printed. If "
        "no purchase order appears on the invoice, return null."
    ),
    "discount_amount": (
        "Return the discount only if a discount line is present (use the printed "
        "value, typically negative). If there is no discount line, return null "
        "rather than 0."
    ),
    "due_date": (
        "Return the payment due date only if it is printed. If the invoice shows "
        "no due date, return null; do not derive one from the invoice date."
    ),
    "tax_amount": (
        "Return the tax/VAT amount only if a tax line is present. If the invoice "
        "shows no tax, return null rather than 0."
    ),
    "invoice_date": (
        "The invoice date is printed in European day/month/year order (e.g. "
        "07/02/2024 means 7 February 2024). Return it in ISO 8601 (YYYY-MM-DD)."
    ),
}


def make_nullable(schema: dict, optional_fields=OPTIONAL_FIELDS) -> dict:
    out = copy.deepcopy(schema)
    props = out.get("properties", {})
    for name in optional_fields:
        prop = props.get(name)
        if prop is None:
            continue
        t = prop.get("type")
        if isinstance(t, str) and t != "null":
            prop["type"] = [t, "null"]
        elif isinstance(t, list) and "null" not in t:
            prop["type"] = t + ["null"]
    # NB: optional fields stay in `required`. Strict structured output emits
    # every declared property regardless, so optionality is carried by the null
    # type; keeping the field required is the recommended strict-mode shape.
    return out


def add_reasoning(schema: dict, prompts=REASONING_PROMPTS, system_prompt=SYSTEM_PROMPT) -> dict:
    out = copy.deepcopy(schema)
    if system_prompt:
        out["X-SystemPrompt"] = system_prompt
    props = out.get("properties", {})
    for name, prompt in prompts.items():
        if name in props:
            props[name]["X-ReasoningPrompt"] = prompt
    return out


def build_all():
    with open(os.path.join(SCHEMAS_DIR, "baseline.json"), encoding="utf-8") as f:
        baseline = json.load(f)
    nullable = make_nullable(baseline)
    reasoning = add_reasoning(nullable)
    for name, sch in (("nullable", nullable), ("reasoning", reasoning)):
        path = os.path.join(SCHEMAS_DIR, f"{name}.json")
        with open(path, "w", encoding="utf-8") as f:
            json.dump(sch, f, indent=2, ensure_ascii=False)
        print("wrote", path)


if __name__ == "__main__":
    build_all()
