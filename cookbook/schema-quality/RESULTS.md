# Schema-Quality Experiment: nullable, reasoning & enum vs the generated baseline

Five invoices of the same type, processed by one schema. They differ in which optional fields are present (customer name, customer code, purchase order, discount, due date) and in *convention*: the discount is printed negative, several dates use European `DD/MM/YYYY`, and the currency is printed in verbose forms (`Euros`, `US Dollars`, `Pounds Sterling`, `€`) while ground truth is the ISO-4217 code. Four schema variants are compared, each adding one lever:

- **baseline** — produced by `retab schemas generate` from one invoice; every field `required`, no nullable types, no reasoning prompts.
- **nullable** — baseline with the optional fields retyped `["<type>", "null"]`; they stay `required` (optionality is carried by the null type, the right shape for strict structured output).
- **reasoning** — nullable plus an `X-ReasoningPrompt` on each optional field (and an `X-SystemPrompt`) telling the model to return null when a field is absent.
- **enum** — reasoning plus an `enum` of ISO-4217 codes on `currency`, normalizing the printed vocabulary to a canonical code.

Each variant runs every invoice at `n_consensus=5`. Likelihood = mean per-field consensus confidence; weak = below 0.90.

## 1. Corpus summary

| Variant | Overall accuracy | Present-field accuracy | **Absent-field accuracy** | Mean likelihood |
|---|---:|---:|---:|---:|
| baseline | 83% (50/60) | 86% (42/49) | **73%** (8/11) | 0.98 |
| nullable | 90% (54/60) | 88% (43/49) | **100%** (11/11) | 1.00 |
| reasoning | 95% (57/60) | 94% (46/49) | **100%** (11/11) | 0.99 |
| enum | 98% (59/60) | 98% (48/49) | **100%** (11/11) | 0.99 |

The **absent-field** column is the headline: it measures how each variant handles fields that are genuinely missing on an invoice — exactly where a required-everything schema must invent a value.

## 2. Accuracy per invoice

| Invoice | Absent fields | baseline | nullable | reasoning | enum |
|---|---|---:|---:|---:|---:|
| invoice_full | — (all present) | 92% | 92% | 100% | 100% |
| invoice_no_name | customer_name | 83% | 83% | 92% | 100% |
| invoice_no_code | customer_code, purchase_order | 75% | 83% | 92% | 92% |
| invoice_minimal | due_date, customer_name, customer_code, purchase_order, discount, tax | 75% | 92% | 100% | 100% |
| invoice_mixed | purchase_order, discount | 92% | 100% | 92% | 100% |

## 3. What happened on the absent fields

For every field that is *absent* on an invoice, the correct answer is `null`. Here is what each variant actually returned.

| Invoice · field | baseline | nullable | reasoning | enum |
|---|---|---|---|---|
| invoice_no_name · customer_name | ✓ `None` | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_no_code · customer_code | ✓ `None` | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_no_code · purchase_order | ✓ `None` | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_minimal · due_date | ✓ `None` | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_minimal · customer_name | ✓ `None` | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_minimal · customer_code | ✓ `None` | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_minimal · purchase_order | ✓ `None` | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_minimal · discount | ✗ `0` | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_minimal · tax | ✗ `0` | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_mixed · purchase_order | ✓ `None` | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_mixed · discount | ✗ `0` | ✓ `None` | ✓ `None` | ✓ `None` |

## 4. Currency normalization (the enum lever)

`currency` is present on every invoice but printed in a verbose form; the correct value is the ISO-4217 code. A free-text field echoes the printed vocabulary, an `enum` normalizes it.

| Invoice (→ expected code) | baseline | nullable | reasoning | enum |
|---|---|---|---|---|
| invoice_full (→ `EUR`) | ✓ `'EUR'` | ✓ `'EUR'` | ✓ `'EUR'` | ✓ `'EUR'` |
| invoice_no_name (→ `EUR`) | ✗ `'Euros'` | ✗ `'Euros'` | ✗ `'Euros'` | ✓ `'EUR'` |
| invoice_no_code (→ `USD`) | ✗ `'US Dollars'` | ✗ `'US Dollars'` | ✗ `'US Dollars'` | ✓ `'USD'` |
| invoice_minimal (→ `GBP`) | ✓ `'GBP'` | ✗ `'Pounds Sterling'` | ✓ `'GBP'` | ✓ `'GBP'` |
| invoice_mixed (→ `EUR`) | ✓ `'EUR'` | ✓ `'EUR'` | ✗ `'€'` | ✓ `'EUR'` |
