# XPO NEW – PRODUCTION — regression tests

Deterministic regression tests for the **"Convert & Merge for XPO"** function
block of the `XPO NEW - PRODUCTION` workflow. This is the block where the
"fix one customer, break another" issues happen, so it's the highest-value
place to guard.

**Run this after any change to the workflow, before publishing.**

## What it checks (core logic)

- **Single-commodity customers** (FilmTec / AGCO / GIMA): a multi-row order is
  collapsed to **one** commodity per stop, with weight / quantity / volume /
  linear-meters **summed**.
- **Every other customer** keeps its **per-row** commodities (not collapsed).
- **Same-customer merge**: several orders for one customer in a single document
  become **one** merged order.
- **Multi-client documents**: different customers stay as **separate** orders and
  their sums are **never mixed**.
- **Customer name → XPO code** resolution via the live lookup table, including
  accent/case tolerance, the single-commodity allow-list, and that every code in
  the allow-list maps to a real customer (catches typos).

## How it stays honest

It contains **no copy** of the function code. On every run it pulls the
**current** code and the **current** customer-codes table straight from the
workflow via the `retab` CLI, then exercises the real functions. So it always
tests what is actually deployed in the **draft** — i.e. your unpublished edits.

It only runs the pure, deterministic helpers (grouping / merging / single-
commodity / customer lookup). It **never** runs the workflow end to end, so it
**never** calls the XPO API. Zero production side effects.

## Prerequisites

- Python 3.9+
- `retab` CLI on PATH, logged in and scoped to the **XPO** organization:

  ```
  retab auth status      # organization.name should be "XPO"
  retab org switch XPO   # if it isn't
  ```

## Usage

```
cd xpo_workflow_tests
python regression_tests.py                 # default workflow (XPO NEW - PRODUCTION)
python regression_tests.py <workflow-id>   # e.g. test the PREPROD workflow
```

Exit code `0` = all passed, `1` = at least one failure.

Typical loop: edit the workflow draft → `python regression_tests.py` → if green,
`retab workflows publish <workflow-id> --description "..."`.

## Extending it

- **New single-commodity customer** (like FilmTec): add its code to
  `SINGLE_COMMODITY_CUSTOMERS` in the function block, then it's already covered —
  the allow-list sanity test will confirm the code maps to a real customer.
- **New scenario**: add a test inside `run_tests()` using the `order_dict(...)`
  and `commodity(...)` builders and the `check(...)` assertion helper.

## Scope / not covered

- **Extraction quality** (the LLM `Extract` block) is *not* covered here — it's
  non-deterministic. Use the **Evals** page in the Retab dashboard (capture a
  representative run and pin expected fields), or `retab workflows experiments`
  on the extract block, for that.
- This harness covers the **deterministic transform logic** only. Address
  truncation, reference codes, pricing totals and hazard-goods passthrough were
  deliberately left out (you chose the "core logic" scope) but can be added with
  the same builders if needed.
