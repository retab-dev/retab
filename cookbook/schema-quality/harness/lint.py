"""A static linter for Retab extraction schemas.

It reads a JSON schema and flags patterns that the experiments in this folder
showed to hurt extraction quality. It makes NO API calls and reads no documents
- it only inspects the schema, so it is free and instant. Like any linter, the
findings are *risks* to judge, not certain bugs.

Each rule traces to a measured failure (see RESULTS.md):

  all-required-no-nullable   The agent baseline marked every field required with
                             no nullable type; on documents missing an optional
                             field the model fabricated a value (discount/tax -> 0).
  required-number-not-null   A required, non-nullable number cannot say "absent",
                             so it is the field most likely to be fabricated as 0.
  sign-convention-no-reason  A discount/credit/adjustment field with no
                             X-ReasoningPrompt: the baseline returned +150 for a
                             -150 discount. A bare type cannot encode the sign.
  enum-candidate             A categorical-looking string with no enum; an enum
                             normalizes output to a known vocabulary.
  missing-description        Fields without a description extract less reliably.

Usage:
    python -m harness.lint schemas/baseline.json
    python -m harness.lint schemas/*.json
"""

from __future__ import annotations

import glob
import json
import sys
from dataclasses import dataclass

SIGN_AMBIGUOUS = ("discount", "adjustment", "credit", "refund", "rebate")
ENUMISH = {"status", "type", "currency", "category", "state", "kind", "priority"}
_SEV_ORDER = {"warn": 0, "info": 1}


@dataclass
class Finding:
    severity: str   # "warn" | "info"
    path: str
    rule: str
    message: str


def _types(prop):
    t = prop.get("type")
    if isinstance(t, list):
        return [x for x in t if x != "null"]
    return [t] if t else []


def _is_nullable(prop):
    t = prop.get("type")
    return t == "null" or (isinstance(t, list) and "null" in t)


def _walk_objects(obj_schema, path=""):
    """Yield (path, name, prop, required_bool) for every property, recursing
    into nested objects, array items, and $defs."""
    props = obj_schema.get("properties", {})
    req = set(obj_schema.get("required", []))
    for name, prop in props.items():
        fp = f"{path}.{name}" if path else name
        yield fp, name, prop, name in req
        if isinstance(prop.get("properties"), dict):
            yield from _walk_objects(prop, fp)
        items = prop.get("items")
        if isinstance(items, dict) and isinstance(items.get("properties"), dict):
            yield from _walk_objects(items, fp + "[]")
    for dname, dschema in (obj_schema.get("$defs") or {}).items():
        if isinstance(dschema.get("properties"), dict):
            yield from _walk_objects(dschema, f"$defs.{dname}")


def lint_schema(schema: dict):
    findings = []
    props = schema.get("properties", {})

    # Schema-level: the "generated, everything required, nothing nullable" smell.
    if props:
        all_required = set(props.keys()) == set(schema.get("required", []))
        any_nullable = any(_is_nullable(p) for p in props.values())
        if all_required and not any_nullable:
            findings.append(Finding(
                "warn", "<schema>", "all-required-no-nullable",
                f"all {len(props)} top-level properties are required and none is "
                f"nullable - optional fields will be fabricated on documents that "
                f"omit them. Mark genuinely-optional fields nullable.",
            ))

    for fp, name, prop, required in _walk_objects(schema):
        types = _types(prop)
        nullable = _is_nullable(prop)
        low = name.lower()
        numeric = "number" in types or "integer" in types

        if required and not nullable and numeric:
            findings.append(Finding(
                "info", fp, "required-number-not-null",
                "required non-nullable number - if this field can be absent on "
                "some documents, extraction will fabricate 0. Consider "
                '["number", "null"].',
            ))

        if numeric and any(k in low for k in SIGN_AMBIGUOUS) \
                and "X-ReasoningPrompt" not in prop:
            findings.append(Finding(
                "warn", fp, "sign-convention-no-reason",
                "discount/credit-type amount with no X-ReasoningPrompt - the "
                "sign convention (these are typically negative) is easy to get "
                "wrong. Add a reasoning prompt stating the expected sign.",
            ))

        if "string" in types and "enum" not in prop and low in ENUMISH:
            findings.append(Finding(
                "info", fp, "enum-candidate",
                f'string "{name}" looks categorical but has no enum - an enum '
                f"normalizes output to a fixed vocabulary.",
            ))

        if not prop.get("description") and not ({"object", "array"} & set(types)):
            findings.append(Finding(
                "info", fp, "missing-description",
                "no description - descriptions measurably improve extraction.",
            ))

    findings.sort(key=lambda f: (_SEV_ORDER.get(f.severity, 9), f.path))
    return findings


def _print(path, findings):
    n_warn = sum(f.severity == "warn" for f in findings)
    n_info = sum(f.severity == "info" for f in findings)
    print(f"\n{path}  ({n_warn} warn, {n_info} info)")
    if not findings:
        print("  clean - no issues found")
        return
    for f in findings:
        tag = "WARN" if f.severity == "warn" else "info"
        print(f"  [{tag}] {f.path}: {f.message}  ({f.rule})")


def main(argv=None):
    argv = argv if argv is not None else sys.argv[1:]
    if not argv:
        print("usage: python -m harness.lint <schema.json> [more.json ...]")
        return 2
    paths = []
    for a in argv:
        paths.extend(glob.glob(a) or [a])
    total_warn = 0
    for p in paths:
        with open(p, encoding="utf-8") as f:
            schema = json.load(f)
        findings = lint_schema(schema)
        total_warn += sum(x.severity == "warn" for x in findings)
        _print(p, findings)
    return 1 if total_warn else 0


if __name__ == "__main__":
    raise SystemExit(main())
