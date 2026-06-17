"""A static linter for Retab extraction schemas.

It reads a JSON schema and flags patterns that hurt extraction quality. It makes
NO API calls and reads no documents — it only inspects the schema, so it is
free, instant, and fully deterministic (no AI/LLM in the loop). Like any linter,
the findings are *risks* to judge, not certain bugs.

The linter is GENERAL: it runs on any Retab / JSON schema, not just invoices.
Its rules come in two kinds:

  Structural rules — domain-agnostic, always on. Judged purely from the schema's
  shape, so they work on any schema:
    no-nullable-fields, non-nullable-number, untyped-field,
    array-without-items, unconstrained-object, missing-description

  Lexical rules — vocabulary-driven, configurable. A field's *semantics* (does
  it carry a sign? is it categorical?) cannot be derived from structure alone, so
  these match the field NAME against a term list. Sensible defaults are provided;
  pass your own to fit your domain (or use --sign-terms / --enum-terms on the CLI):
    sign-convention-no-reason, enum-candidate

No rule ever looks at whether a field is in `required`: a present field is
extracted regardless of `required`, and under strict structured output every
property is emitted anyway — so nullability (the type), not `required`, is what
lets the model report "absent".

See ../LINTER.md for a full, per-rule explanation, and ../RESULTS.md for the
measurements behind the rules that are backed by the experiment.

Usage:
    python -m harness.lint schemas/baseline.json
    python -m harness.lint schemas/*.json
    python -m harness.lint my.json --sign-terms discount,markdown --enum-terms status,region
"""

from __future__ import annotations

import argparse
import glob
import json
import sys
from dataclasses import dataclass

# Default vocabularies for the two lexical rules. These are starting points, not
# truths — override them per domain on the CLI or via lint_schema(...).
DEFAULT_SIGN_TERMS = (
    "discount", "adjustment", "credit", "refund", "rebate", "markdown",
    "chargeback", "delta", "variance",
)
DEFAULT_ENUM_TERMS = frozenset({
    "status", "type", "currency", "category", "state", "kind", "priority",
    "mode", "level", "method", "unit", "country", "language", "region",
})
_SEV_ORDER = {"warn": 0, "info": 1}


@dataclass
class Finding:
    severity: str   # "warn" | "info"
    path: str
    rule: str
    message: str


def _types(prop):
    """The declared type(s) of a property, with "null" stripped out."""
    t = prop.get("type")
    if isinstance(t, list):
        return [x for x in t if x != "null"]
    return [t] if t else []


def _is_nullable(prop):
    """True if the property's type admits null."""
    t = prop.get("type")
    return t == "null" or (isinstance(t, list) and "null" in t)


def _has_any_type(prop):
    """True if the property constrains its value at all (a type, an enum, or a
    combinator/ref). A property with none of these is under-specified."""
    return bool(prop.get("type")) or any(
        k in prop for k in ("enum", "const", "$ref", "anyOf", "oneOf", "allOf")
    )


def _walk_objects(obj_schema, path=""):
    """Yield (path, name, prop) for every property, recursing into nested
    objects, array items, and $defs. `required` is intentionally ignored — the
    rules judge a field by its type, not its membership in `required`."""
    props = obj_schema.get("properties", {})
    for name, prop in props.items():
        fp = f"{path}.{name}" if path else name
        yield fp, name, prop
        if isinstance(prop.get("properties"), dict):
            yield from _walk_objects(prop, fp)
        items = prop.get("items")
        if isinstance(items, dict) and isinstance(items.get("properties"), dict):
            yield from _walk_objects(items, fp + "[]")
    for dname, dschema in (obj_schema.get("$defs") or {}).items():
        if isinstance(dschema.get("properties"), dict):
            yield from _walk_objects(dschema, f"$defs.{dname}")


def lint_schema(schema: dict, sign_terms=DEFAULT_SIGN_TERMS, enum_terms=DEFAULT_ENUM_TERMS):
    """Return a list of Finding for `schema`.

    sign_terms / enum_terms drive the two lexical rules; override them to fit a
    non-invoice domain. Everything else is structural and domain-agnostic.
    """
    findings = []
    enum_terms = {t.lower() for t in enum_terms}
    sign_terms = tuple(t.lower() for t in sign_terms)
    props = schema.get("properties", {})

    # --- Schema-level structural rule ---------------------------------------
    # If no field anywhere can be null, the model has no way to say "absent".
    if props:
        any_nullable = any(_is_nullable(p) for _, _, p in _walk_objects(schema))
        if not any_nullable:
            findings.append(Finding(
                "warn", "<schema>", "no-nullable-fields",
                f"no field in this schema is nullable - optional fields will be "
                f"fabricated on documents that omit them. Mark genuinely-optional "
                f"fields nullable, e.g. \"type\": [\"string\", \"null\"].",
            ))

    # --- Per-field rules -----------------------------------------------------
    for fp, name, prop in _walk_objects(schema):
        types = _types(prop)
        nullable = _is_nullable(prop)
        low = name.lower()
        numeric = "number" in types or "integer" in types

        # Structural: a non-nullable number can't say "absent" -> fabricates 0.
        if not nullable and numeric:
            findings.append(Finding(
                "info", fp, "non-nullable-number",
                "non-nullable number - if this field can be absent on some "
                "documents, extraction will fabricate 0. Consider "
                '["number", "null"].',
            ))

        # Structural: a property with no type/enum/$ref is unconstrained.
        if not _has_any_type(prop):
            findings.append(Finding(
                "info", fp, "untyped-field",
                "no type, enum, or $ref - the model is unconstrained on this "
                "field. Declare a concrete type.",
            ))

        # Structural: an array with no items schema is unconstrained.
        if "array" in types and not isinstance(prop.get("items"), dict):
            findings.append(Finding(
                "info", fp, "array-without-items",
                "array with no `items` schema - element shape is unconstrained. "
                "Add an items schema.",
            ))

        # Structural: an object with no properties (and not closed) is free-form.
        if "object" in types and not isinstance(prop.get("properties"), dict):
            findings.append(Finding(
                "info", fp, "unconstrained-object",
                "object with no `properties` - the model may invent arbitrary "
                "keys. Declare the properties.",
            ))

        # Lexical (configurable): a sign-ambiguous numeric field with no reasoning.
        if numeric and any(k in low for k in sign_terms) \
                and "X-ReasoningPrompt" not in prop:
            findings.append(Finding(
                "warn", fp, "sign-convention-no-reason",
                "sign-ambiguous amount with no X-ReasoningPrompt - the sign "
                "convention (these are often negative) is easy to get wrong. Add "
                "a reasoning prompt stating the expected sign.",
            ))

        # Lexical (configurable): a categorical-looking string with no enum.
        if "string" in types and "enum" not in prop and low in enum_terms:
            findings.append(Finding(
                "warn", fp, "enum-candidate",
                f'string "{name}" looks categorical but has no enum - a free-text '
                f"field echoes whatever is printed; an enum normalizes output to a "
                f"fixed vocabulary.",
            ))

        # Structural: missing description (skip container types).
        if not prop.get("description") and not ({"object", "array"} & set(types)):
            findings.append(Finding(
                "info", fp, "missing-description",
                "no description - descriptions are generally recommended "
                "(heuristic; not measured in this folder's experiments).",
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


def _split_terms(s):
    return [t.strip() for t in s.split(",") if t.strip()] if s else None


def main(argv=None):
    ap = argparse.ArgumentParser(
        prog="python -m harness.lint",
        description="Static linter for Retab extraction schemas (no API calls).",
    )
    ap.add_argument("paths", nargs="*", help="schema .json files or globs")
    ap.add_argument("--sign-terms", help="comma-separated field-name terms for the sign rule")
    ap.add_argument("--enum-terms", help="comma-separated field-name terms for the enum rule")
    args = ap.parse_args(argv)

    if not args.paths:
        ap.print_usage()
        return 2

    kwargs = {}
    if args.sign_terms is not None:
        kwargs["sign_terms"] = _split_terms(args.sign_terms)
    if args.enum_terms is not None:
        kwargs["enum_terms"] = _split_terms(args.enum_terms)

    paths = []
    for a in args.paths:
        paths.extend(glob.glob(a) or [a])

    total_warn = 0
    for p in paths:
        with open(p, encoding="utf-8") as f:
            schema = json.load(f)
        findings = lint_schema(schema, **kwargs)
        total_warn += sum(x.severity == "warn" for x in findings)
        _print(p, findings)
    return 1 if total_warn else 0


if __name__ == "__main__":
    raise SystemExit(main())
