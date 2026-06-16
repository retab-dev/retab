# Schema-Quality Experiment: nullable & reasoning vs the generated baseline

Five invoices of the same type, processed by one schema. They differ only in which optional fields are present (customer name, customer code, purchase order, discount, due date). Three schema variants are compared, each adding one lever:

- **baseline** — produced by `retab schemas generate` from one invoice; every field `required`, no nullable types, no reasoning prompts.
- **nullable** — baseline with the optional fields retyped `["<type>", "null"]` and removed from `required`.
- **reasoning** — nullable plus an `X-ReasoningPrompt` on each optional field (and an `X-SystemPrompt`) telling the model to return null when a field is absent.

Each variant runs every invoice at `n_consensus=5`. Likelihood = mean per-field consensus confidence; weak = below 0.90.

## 1. Corpus summary

| Variant | Overall accuracy | Present-field accuracy | **Absent-field accuracy** | Mean likelihood |
|---|---:|---:|---:|---:|
| baseline | 90% (54/60) | 94% (46/49) | **73%** (8/11) | 0.99 |
| nullable | 95% (57/60) | 94% (46/49) | **100%** (11/11) | 1.00 |
| reasoning | 97% (58/60) | 96% (47/49) | **100%** (11/11) | 0.99 |

The **absent-field** column is the headline: it measures how each variant handles fields that are genuinely missing on an invoice — exactly where a required-everything schema must invent a value.

## 2. Accuracy per invoice

| Invoice | Absent fields | baseline | nullable | reasoning |
|---|---|---:|---:|---:|
| invoice_full | — (all present) | 92% | 92% | 92% |
| invoice_no_name | customer_name | 92% | 92% | 100% |
| invoice_no_code | customer_code, purchase_order | 92% | 92% | 92% |
| invoice_minimal | due_date, customer_name, customer_code, purchase_order, discount, tax | 83% | 100% | 100% |
| invoice_mixed | purchase_order, discount | 92% | 100% | 100% |

## 3. What happened on the absent fields

For every field that is *absent* on an invoice, the correct answer is `null`. Here is what each variant actually returned.

| Invoice · field | baseline | nullable | reasoning |
|---|---|---|---|
| invoice_no_name · customer_name | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_no_code · customer_code | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_no_code · purchase_order | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_minimal · due_date | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_minimal · customer_name | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_minimal · customer_code | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_minimal · purchase_order | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_minimal · discount | ✗ `0` | ✓ `None` | ✓ `None` |
| invoice_minimal · tax | ✗ `0` | ✓ `None` | ✓ `None` |
| invoice_mixed · purchase_order | ✓ `None` | ✓ `None` | ✓ `None` |
| invoice_mixed · discount | ✗ `0` | ✓ `None` | ✓ `None` |
