# Schema-Quality Experiment: nullable, reasoning & enum vs the generated baseline

Five invoices of the same type, processed by one schema. They differ in which optional fields are present (customer name, customer code, purchase order, discount, due date) and in *convention*: the discount is printed negative, several dates use European `DD/MM/YYYY`, and the currency is printed in verbose forms (`Euros`, `US Dollars`, `Pounds Sterling`, `€`) while ground truth is the ISO-4217 code. Four schema variants are compared, each adding one lever:

- **baseline** — produced by `retab schemas generate` from one invoice; no nullable types, no reasoning prompts.
- **nullable** — baseline with the optional fields retyped `["<type>", "null"]`, so the model can report "absent" instead of fabricating a value.
- **reasoning** — nullable plus an `X-ReasoningPrompt` on each optional field (and an `X-SystemPrompt`) telling the model to return null when a field is absent.
- **enum** — reasoning plus an `enum` of ISO-4217 codes on `currency`, normalizing the printed vocabulary to a canonical code.

Each variant runs every invoice at `n_consensus=5`. Likelihood = mean per-field consensus confidence (shown as a percentage); weak = below 90%.

## 1. Corpus summary

| Variant | Overall accuracy | Present-field accuracy | **Absent-field accuracy** | Mean likelihood |
|---|---:|---:|---:|---:|
| baseline | 83% (50/60) | 86% (42/49) | **73%** (8/11) | 97.6% |
| nullable | 90% (54/60) | 88% (43/49) | **100%** (11/11) | 99.8% |
| reasoning | 95% (57/60) | 94% (46/49) | **100%** (11/11) | 99.4% |
| enum | 98% (59/60) | 98% (48/49) | **100%** (11/11) | 99.3% |

The **absent-field** column is the headline: it measures how each variant handles fields that are genuinely missing on an invoice — exactly where a schema with no nullable types must invent a value.

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

## 5. Stability — consensus agreement

Likelihood measures how strongly the consensus runs **agreed** on a field, not whether the value was right. A field below the threshold (90%) is **weak** — the runs split. The interesting case is high likelihood on a *wrong* value: the model is confidently wrong, so likelihood cannot be used to rank schemas — only accuracy can.

| Variant | Mean likelihood | Weak fields (< threshold) |
|---|---:|---:|
| baseline | 97.6% | 5 |
| nullable | 99.8% | 1 |
| reasoning | 99.4% | 1 |
| enum | 99.3% | 1 |

Every field that fell below the threshold, with the value returned and whether it was correct:

| Variant | Invoice · field | Likelihood | Returned | Correct |
|---|---|---:|---|:--:|
| baseline | invoice_no_name · customer_name | 60.0% | `None` | ✓ |
| baseline | invoice_no_code · currency | 86.0% | `'US Dollars'` | ✗ |
| baseline | invoice_minimal · due_date | 60.0% | `None` | ✓ |
| baseline | invoice_minimal · currency | 86.0% | `'GBP'` | ✓ |
| baseline | invoice_mixed · currency | 65.3% | `'EUR'` | ✓ |
| nullable | invoice_no_code · currency | 86.0% | `'US Dollars'` | ✗ |
| reasoning | invoice_mixed · currency | 65.3% | `'€'` | ✗ |
| enum | invoice_no_code · discount | 60.0% | `32.25` | ✗ |

Note how few fields are weak even though accuracy varies widely: likelihood stayed near `100%` across variants while accuracy moved **83% → 98%**. A confident value is not a correct one.

### Where stability rose: `currency`

Mean consensus likelihood on the `currency` field, by variant:

| baseline | nullable | reasoning | enum |
|---:|---:|---:|---:|
| 87.5% | 97.2% | 93.1% | 100.0% |

Agreement climbs from `87.5%` to a perfect `100%`. The reason is instructive: `currency` is an **identical free-text field** in baseline, nullable and reasoning — so the wobble between those three (`65%`–`100%`, depending on the document and the run) is pure consensus sampling noise. A free-text field lets each of the five runs choose a different surface form (`€`, `EUR`, `Euros`); when they disagree, likelihood drops. The **enum** is the one change that constrains the decode to a fixed set of codes, so all five runs land on the same token and likelihood locks to `100%` on every invoice. Constraining the output space removes the degrees of freedom the runs were splitting over — the enum raised accuracy **and** stability at once.
