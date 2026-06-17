# Consensus Cost Frontier

Consensus (`n_consensus = K`) runs a document through K independent model calls and merges them. Cost scales with K — so when is it worth it, and does that depend on how good the schema already is? We sweep K over two schemas from the [schema-quality](../schema-quality) cookbook on the same 5 invoices, model `retab-micro`:

- **baseline** — the weak generated schema (no nullable types, no reasoning, no enum).
- **enum** — the strong schema (nullable + reasoning + enum), which scored 98% at K=5.

## Frontier — accuracy & likelihood vs K

| K | baseline accuracy | baseline likelihood | enum accuracy | enum likelihood |
|---:|---:|---:|---:|---:|
| 1 | 73.3% | n/a | 100.0% | n/a |
| 3 | 83.3% | 99.0% | 100.0% | 98.9% |
| 5 | 83.3% | 97.6% | 98.3% | 99.3% |
| 7 | 85.0% | 98.9% | 100.0% | 99.5% |

Accuracy and likelihood are averaged across the 5 invoices. Likelihood is undefined at K=1 (a single call has nothing to agree with).

## Cost vs K

Cost scales **linearly** with K. On `retab-micro` the observed rate is ~**0.20 credits per consensus call**, so extracting one document costs ~`0.20 x K`:

| K | credits / document | vs K=1 |
|---:|---:|---:|
| 1 | 0.20 | 1x |
| 3 | 0.60 | 3x |
| 5 | 1.00 | 5x |
| 7 | 1.40 | 7x |

Wall-clock **latency**, by contrast, stayed roughly flat at ~6-7s across K (the K calls run in parallel, so you pay K times the credits but not K times the wall-clock). Latency is environment-dependent — treat it as indicative.

These are nominal, uncached per-document costs. Actual charges can be lower on repeats, since Retab caches identical (document, schema, model, K) requests server-side.

## What the frontier shows

- **Consensus gives a weak schema only a modest, capped lift.** The baseline rises from ~73% (K=1) to ~83% (K=3) as consensus averages out single-call noise on borderline fields, then plateaus. It never approaches the strong schema, because its remaining errors are *structural* — a non-nullable number can only return `0` for an absent field, a free-text currency can only echo the printed form — and every one of the K calls makes the same mistake, so merging them cannot fix it.
- **A strong schema is saturated at K=1.** The enum schema is ~100% from a single call; extra calls add cost, not accuracy (the K=5 dip is sampling noise on one field).
- **Schema design beats consensus spend.** enum at **K=1 (100%, ~0.2 credits/doc)** beats baseline at **K=7 (85%, ~1.4 credits/doc)** — higher accuracy at one-seventh the cost.
- **What consensus does buy is the confidence signal.** Likelihood exists only for K≥2; it is the per-field agreement used to flag shaky extractions — the real reason to pay for K>1, not an accuracy bump.
- **Cost is linear, latency is flat.** Credits scale ~`0.2 x K` per document; the K calls run in parallel, so wall-clock stays roughly constant.

*Takeaway: fix the schema first — it works at K=1 and is the cheapest path to accuracy. Use a small K (e.g. 3) for a stability signal, not as a substitute for schema design.*

## Per-invoice accuracy by K

**baseline**

| Invoice | K=1 | K=3 | K=5 | K=7 |
|---|---:|---:|---:|---:|
| invoice_full | 91.7% | 91.7% | 91.7% | 91.7% |
| invoice_no_name | 75.0% | 75.0% | 83.3% | 75.0% |
| invoice_no_code | 66.7% | 83.3% | 75.0% | 83.3% |
| invoice_minimal | 50.0% | 75.0% | 75.0% | 83.3% |
| invoice_mixed | 83.3% | 91.7% | 91.7% | 91.7% |

**enum**

| Invoice | K=1 | K=3 | K=5 | K=7 |
|---|---:|---:|---:|---:|
| invoice_full | 100.0% | 100.0% | 100.0% | 100.0% |
| invoice_no_name | 100.0% | 100.0% | 100.0% | 100.0% |
| invoice_no_code | 100.0% | 100.0% | 91.7% | 100.0% |
| invoice_minimal | 100.0% | 100.0% | 100.0% | 100.0% |
| invoice_mixed | 100.0% | 100.0% | 100.0% | 100.0% |
