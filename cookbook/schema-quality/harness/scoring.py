"""Score an extraction record against the corpus ground truth.

Scoring is typed and normalized so that cosmetic differences are not counted as
errors:
  - numbers      : coerced to float, compared with absolute + relative tolerance
  - dates        : both sides parsed to a date before comparison
  - strings      : trimmed, whitespace-collapsed, casefolded; IBAN-style values
                   also compared with internal spaces removed
  - null         : a ground-truth value of None is correct iff the field is
                   absent or null in the output (the nullable lesson)
  - "<field>_count" ground-truth keys are resolved by counting "<field>" in the
                   output array (e.g. line_items_count -> len(output.line_items))

Each scored field yields: correct (bool), expected, got, and likelihood (the
per-field consensus confidence, when available).
"""

from __future__ import annotations

import datetime as _dt
import re
from dataclasses import dataclass
from typing import Any, Optional

_MISSING = object()
_NUM_RE = re.compile(r"[^0-9.\-]")
_DATE_FORMATS = ("%Y-%m-%d", "%m/%d/%Y", "%d/%m/%Y", "%B %d, %Y", "%b %d, %Y", "%d %B %Y")


@dataclass
class FieldScore:
    field: str
    correct: bool
    expected: Any
    got: Any
    likelihood: Optional[float]


@dataclass
class DocScore:
    doc: str
    variant: str
    fields: list
    credits: Optional[int] = None

    @property
    def accuracy(self) -> float:
        return sum(f.correct for f in self.fields) / len(self.fields) if self.fields else 0.0

    @property
    def mean_likelihood(self) -> Optional[float]:
        vals = [f.likelihood for f in self.fields if f.likelihood is not None]
        return sum(vals) / len(vals) if vals else None

    def weak_fields(self, threshold: float) -> list:
        return [f.field for f in self.fields
                if f.likelihood is not None and f.likelihood < threshold]


def _to_float(v: Any) -> Optional[float]:
    if isinstance(v, bool):
        return None
    if isinstance(v, (int, float)):
        return float(v)
    if isinstance(v, str):
        s = v.strip()
        neg = s.startswith("(") and s.endswith(")")
        s = _NUM_RE.sub("", s)
        if s in ("", "-", "."):
            return None
        try:
            f = float(s)
            return -f if neg else f
        except ValueError:
            return None
    return None


def _to_date(v: Any) -> Optional[_dt.date]:
    if not isinstance(v, str):
        return None
    s = v.strip()
    for fmt in _DATE_FORMATS:
        try:
            return _dt.datetime.strptime(s, fmt).date()
        except ValueError:
            continue
    return None


def _norm_str(v: Any) -> str:
    return re.sub(r"\s+", " ", str(v).strip()).casefold()


def _scalar_likelihood(v: Any) -> Optional[float]:
    """Reduce a likelihood to a scalar.

    Array and nested-object fields report likelihoods as lists/dicts of
    per-element confidences; collapse those to their mean so a field has a
    single stability number.
    """
    if v is None:
        return None
    if isinstance(v, bool):
        return float(v)
    if isinstance(v, (int, float)):
        return float(v)
    if isinstance(v, dict):
        v = list(v.values())
    if isinstance(v, list):
        leaves = [_scalar_likelihood(x) for x in v]
        leaves = [x for x in leaves if x is not None]
        return sum(leaves) / len(leaves) if leaves else None
    return None


def _values_match(expected: Any, got: Any) -> bool:
    if got is _MISSING:
        return expected is None
    if expected is None:
        return got is None or got is _MISSING
    if got is None:
        return False

    if isinstance(expected, bool):
        return bool(got) == expected

    if isinstance(expected, (int, float)):
        gf = _to_float(got)
        if gf is None:
            return False
        return abs(gf - float(expected)) <= max(0.01, 1e-3 * abs(float(expected)))

    # string-ish: try date, then normalized string, then space-stripped string
    ed, gd = _to_date(expected), _to_date(got)
    if ed and gd:
        return ed == gd
    ne, ng = _norm_str(expected), _norm_str(got)
    if ne == ng:
        return True
    return ne.replace(" ", "") == ng.replace(" ", "")


def _get_path(obj: Any, path: str):
    """Resolve a dotted path (e.g. 'bill_to.name') against nested dicts."""
    cur = obj
    for part in path.split("."):
        if isinstance(cur, dict) and part in cur:
            cur = cur[part]
        else:
            return _MISSING
    return cur


def score(record: dict, ground_truth: dict, field_map: dict,
          doc: str, variant: str) -> DocScore:
    """Score a record. ground_truth uses canonical field names; field_map maps
    each canonical name to the property path actually used by the schema."""
    output = record.get("output") or {}
    likelihoods = (record.get("consensus") or {}).get("likelihoods") or {}
    credits = (record.get("usage") or {}).get("credits")

    fields = []
    for key, expected in ground_truth.items():
        path = field_map.get(key, key)
        got = _get_path(output, path)
        correct = _values_match(expected, got)
        lk_raw = _get_path(likelihoods, path)
        likelihood = _scalar_likelihood(None if lk_raw is _MISSING else lk_raw)
        fields.append(FieldScore(
            field=key,
            correct=correct,
            expected=expected,
            got=None if got is _MISSING else got,
            likelihood=likelihood,
        ))
    return DocScore(doc=doc, variant=variant, fields=fields, credits=credits)
