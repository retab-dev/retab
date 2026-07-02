# Consensus Cost Frontier

A measured sequel to the [schema-quality](../schema-quality) cookbook. That one
showed *what* to put in a schema; this one asks a cost question: **consensus
(`n_consensus = K`) runs each document through K independent model calls and
merges them — how much does that buy, and is it worth the linear cost?**

## The question

Consensus improves reliability, but credits scale with K. So:
- Does more consensus raise **accuracy**?
- Does the answer depend on how good the **schema** already is?
- What is the actual **cost** (credits) and **latency** curve?

## Method

One corpus (the same 5 invoices), one model (`retab-micro`), and **two schemas**
swept across `K ∈ {1, 3, 5, 7}` so the only variables are K and schema quality:

- **baseline** — the weak generated schema (no nullable types, no reasoning, no enum).
- **enum** — the strong schema (nullable + reasoning + enum) from schema-quality.

For each `(schema, invoice, K)` we extract via the Retab CLI, score against the
ground-truth manifest, and record **accuracy**, **consensus likelihood**,
**credits** (`usage.credits`), and **wall-clock latency**.

This reuses the schema-quality corpus, schemas, and harness (runner + scoring),
including its on-disk cache — so `K=5` is reused for free and only the new K
values cost credits. `retab-micro` keeps the spend small.

## Results

Full output in [`RESULTS.md`](RESULTS.md). Headline:

| K | baseline accuracy | enum accuracy | credits / document |
|---:|---:|---:|---:|
| 1 | 73.3% | **100.0%** | 0.20 |
| 3 | 83.3% | 100.0% | 0.60 |
| 5 | 83.3% | 98.3%¹ | 1.00 |
| 7 | 85.0% | 100.0% | 1.40 |

¹ the K=5 dip is sampling noise on one field, not a K effect. Cost is the
nominal per-document figure (~0.20 × K on `retab-micro`); it scales linearly
with K while accuracy plateaus.

### Interpretation

- **Consensus gives a weak schema only a modest, capped lift** (baseline 73% → 83%
  from K=1→K=3, then plateau). Its remaining errors are *structural* — every one
  of the K calls makes the same mistake, so merging cannot fix it.
- **A strong schema is saturated at K=1** — extra calls add cost, not accuracy.
- **Schema design beats consensus spend:** enum at **K=1 (100%, ~0.2 credits/doc)**
  beats baseline at **K=7 (85%, ~1.4 credits/doc)** — higher accuracy at
  one-seventh the cost.
- **What consensus buys is the confidence signal** (likelihood, defined only for
  K≥2), used to flag shaky fields — not an accuracy bump.
- **Cost is linear (~0.2 × K credits/doc), latency is flat** (~6–7s; the K calls
  run in parallel).

*Takeaway: fix the schema first — it works at K=1 and is the cheapest path to
accuracy. Use a small K (e.g. 3) for a stability signal, not as a substitute for
schema design.*

## How to run

```bash
# Requires the sibling schema-quality cookbook (corpus, schemas, harness).
# Build its corpus/schemas first if you haven't:
#   (cd ../schema-quality && python generate_documents.py && python -m harness.variants)

python sweep.py                       # writes RESULTS.md (reuses cache; only new K cost credits)
python sweep.py --n 1,2,3,5,7         # custom sweep
python sweep.py --schemas enum        # one schema only
```

A narrated walkthrough is in [`consensus-cost.ipynb`](consensus-cost.ipynb).

## File guide

| Path | Role |
|---|---|
| `sweep.py` | The experiment: sweeps schemas × K, scores, writes `RESULTS.md`. |
| `RESULTS.md` | Latest measured output. |
| `consensus-cost.ipynb` | Narrated, runnable walkthrough. |

## Relation to schema-quality

This cookbook **depends on** [`../schema-quality`](../schema-quality): it imports
its `harness` (runner + scoring) and reads its corpus and schemas. The two are
complementary — schema-quality shows which schema design choices matter;
consensus-cost shows that, once the schema is right, consensus is for confidence
rather than accuracy.
