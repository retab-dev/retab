"""Render the 3-variant experiment results as Markdown.

results is a dict: {doc_stem: {variant: DocScore}}.
The headline metric is accuracy on *absent* fields (ground truth = null), because
that is where a "everything required" baseline is forced to hallucinate.
"""

from __future__ import annotations

from typing import Optional

VARIANTS = ["baseline", "nullable", "reasoning", "enum"]


def _pct(num, den):
    return f"{100 * num / den:.0f}%" if den else "—"


def _lik(x: Optional[float]):
    return "n/a" if x is None else f"{x:.2f}"


def _agg(scores):
    """Aggregate a list of DocScore for one variant across the corpus."""
    all_fields = [f for ds in scores for f in ds.fields]
    present = [f for f in all_fields if f.expected is not None]
    absent = [f for f in all_fields if f.expected is None]
    liks = [f.likelihood for f in all_fields if f.likelihood is not None]
    return {
        "overall": (sum(f.correct for f in all_fields), len(all_fields)),
        "present": (sum(f.correct for f in present), len(present)),
        "absent": (sum(f.correct for f in absent), len(absent)),
        "mean_lik": sum(liks) / len(liks) if liks else None,
    }


def render(results: dict, threshold: float) -> str:
    docs = list(results.keys())
    L = []
    L.append("# Schema-Quality Experiment: nullable, reasoning & enum vs the generated baseline\n")
    L.append(
        "Five invoices of the same type, processed by one schema. They differ "
        "in which optional fields are present (customer name, customer code, "
        "purchase order, discount, due date) and in *convention*: the discount "
        "is printed negative, several dates use European `DD/MM/YYYY`, and the "
        "currency is printed in verbose forms (`Euros`, `US Dollars`, `Pounds "
        "Sterling`, `€`) while ground truth is the ISO-4217 code. Four schema "
        "variants are compared, each adding one lever:\n"
    )
    L.append("- **baseline** — produced by `retab schemas generate` from one invoice; every field `required`, no nullable types, no reasoning prompts.")
    L.append("- **nullable** — baseline with the optional fields retyped `[\"<type>\", \"null\"]`; they stay `required` (optionality is carried by the null type, the right shape for strict structured output).")
    L.append("- **reasoning** — nullable plus an `X-ReasoningPrompt` on each optional field (and an `X-SystemPrompt`) telling the model to return null when a field is absent.")
    L.append("- **enum** — reasoning plus an `enum` of ISO-4217 codes on `currency`, normalizing the printed vocabulary to a canonical code.\n")
    L.append(f"Each variant runs every invoice at `n_consensus=5`. Likelihood = mean per-field consensus confidence; weak = below {threshold:.2f}.\n")

    # 1. Corpus summary
    L.append("## 1. Corpus summary\n")
    L.append("| Variant | Overall accuracy | Present-field accuracy | **Absent-field accuracy** | Mean likelihood |")
    L.append("|---|---:|---:|---:|---:|")
    aggs = {}
    for v in VARIANTS:
        scores = [results[d][v] for d in docs if v in results[d]]
        a = _agg(scores)
        aggs[v] = a
        L.append(
            f"| {v} | {_pct(*a['overall'])} ({a['overall'][0]}/{a['overall'][1]}) "
            f"| {_pct(*a['present'])} ({a['present'][0]}/{a['present'][1]}) "
            f"| **{_pct(*a['absent'])}** ({a['absent'][0]}/{a['absent'][1]}) "
            f"| {_lik(a['mean_lik'])} |"
        )
    L.append("")
    L.append(
        "The **absent-field** column is the headline: it measures how each "
        "variant handles fields that are genuinely missing on an invoice — "
        "exactly where a required-everything schema must invent a value.\n"
    )

    # 2. Per-document accuracy
    head = " | ".join(VARIANTS)
    align = "|".join(["---:"] * len(VARIANTS))
    L.append("## 2. Accuracy per invoice\n")
    L.append(f"| Invoice | Absent fields | {head} |")
    L.append(f"|---|---|{align}|")
    for d in docs:
        absent = [f.field for f in results[d]["baseline"].fields if f.expected is None]
        cells = []
        for v in VARIANTS:
            ds = results[d][v]
            cells.append(_pct(sum(f.correct for f in ds.fields), len(ds.fields)))
        absent_str = ", ".join(absent) if absent else "— (all present)"
        L.append(f"| {d} | {absent_str} | " + " | ".join(cells) + " |")
    L.append("")

    # 3. What each variant did on the absent fields
    cell_align = "|".join(["---"] * len(VARIANTS))
    L.append("## 3. What happened on the absent fields\n")
    L.append("For every field that is *absent* on an invoice, the correct answer is `null`. Here is what each variant actually returned.\n")
    L.append(f"| Invoice · field | {head} |")
    L.append(f"|---|{cell_align}|")
    any_row = False
    for d in docs:
        base_fields = {f.field: f for f in results[d]["baseline"].fields}
        for field, bf in base_fields.items():
            if bf.expected is not None:
                continue
            any_row = True
            cells = []
            for v in VARIANTS:
                fs = {f.field: f for f in results[d][v].fields}[field]
                mark = "✓" if fs.correct else "✗"
                cells.append(f"{mark} `{fs.got!r}`")
            L.append(f"| {d} · {field} | " + " | ".join(cells) + " |")
    if not any_row:
        L.append("| (no absent fields in corpus) | | | |")
    L.append("")

    # 4. Currency normalization (the enum lever)
    L.append("## 4. Currency normalization (the enum lever)\n")
    L.append("`currency` is present on every invoice but printed in a verbose form; the correct value is the ISO-4217 code. A free-text field echoes the printed vocabulary, an `enum` normalizes it.\n")
    L.append(f"| Invoice (→ expected code) | {head} |")
    L.append(f"|---|{cell_align}|")
    for d in docs:
        cf = {f.field: f for f in results[d]["baseline"].fields}.get("currency")
        if cf is None:
            continue
        cells = []
        for v in VARIANTS:
            fs = {f.field: f for f in results[d][v].fields}["currency"]
            mark = "✓" if fs.correct else "✗"
            cells.append(f"{mark} `{fs.got!r}`")
        L.append(f"| {d} (→ `{cf.expected}`) | " + " | ".join(cells) + " |")
    L.append("")

    return "\n".join(L)
