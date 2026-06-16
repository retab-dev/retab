"""Run the 3-variant invoice experiment.

One schema set (baseline / nullable / reasoning) is applied to all five
invoices. For each (invoice, variant) we extract at n_consensus=K via the Retab
CLI, score against documents/manifest.json using schemas/field_map.json, and
render a Markdown report.

Usage:
    python -m harness.experiment                 # all docs, all variants (cached)
    python -m harness.experiment --no-cache      # force fresh extractions
    python -m harness.experiment --n-consensus 5 --model retab-micro
"""

from __future__ import annotations

import argparse
import json
import os
import sys

from .report import VARIANTS, render
from .runner import run_extraction
from .scoring import score

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DOCS_DIR = os.path.join(ROOT, "documents")
SCHEMAS_DIR = os.path.join(ROOT, "schemas")


def _load_json(path):
    with open(path, encoding="utf-8") as f:
        return json.load(f)


def _stem(doc_file):
    return os.path.splitext(doc_file)[0]


def _load_field_map():
    fm = _load_json(os.path.join(SCHEMAS_DIR, "field_map.json"))
    return {k: v for k, v in fm.items() if not k.startswith("_")}


def run(model="retab-micro", n_consensus=5, use_cache=True, threshold=0.9):
    manifest = _load_json(os.path.join(DOCS_DIR, "manifest.json"))
    field_map = _load_field_map()
    schemas = {v: _load_json(os.path.join(SCHEMAS_DIR, f"{v}.json")) for v in VARIANTS}

    results = {}
    for entry in manifest["documents"]:
        stem = _stem(entry["file"])
        doc_path = os.path.join(DOCS_DIR, entry["file"])
        gt = entry["ground_truth"]
        results[stem] = {}
        for v in VARIANTS:
            print(f"[{stem}] {v} ...", file=sys.stderr, flush=True)
            rec = run_extraction(doc_path, schemas[v], model, n_consensus, use_cache)
            ds = score(rec, gt, field_map, stem, v)
            cached = " (cached)" if rec.get("_cached") else ""
            print(f"[{stem}] {v} acc={ds.accuracy:.0%}{cached}",
                  file=sys.stderr, flush=True)
            results[stem][v] = ds

    return render(results, threshold)


def main():
    ap = argparse.ArgumentParser(description="Run the 3-variant invoice experiment.")
    ap.add_argument("--model", default="retab-micro")
    ap.add_argument("--n-consensus", type=int, default=5)
    ap.add_argument("--threshold", type=float, default=0.9)
    ap.add_argument("--no-cache", action="store_true")
    ap.add_argument("--out", default=os.path.join(ROOT, "RESULTS.md"))
    args = ap.parse_args()

    md = run(
        model=args.model,
        n_consensus=args.n_consensus,
        use_cache=not args.no_cache,
        threshold=args.threshold,
    )
    with open(args.out, "w", encoding="utf-8") as f:
        f.write(md)
    print(f"\nWrote report to {args.out}", file=sys.stderr)


if __name__ == "__main__":
    main()
