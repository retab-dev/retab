# The schema linter — how it works

`harness/lint.py` is a **static linter for Retab extraction schemas**. It reads a
JSON Schema and flags design patterns that hurt extraction quality.

- **No AI, no network, no documents.** It is a plain, deterministic algorithm
  over the schema's JSON — given the same schema it always returns the same
  findings, instantly and for free. (An AI *wrote* it; what *runs* is pure code.)
- **General.** It runs on any Retab / JSON schema, not just the invoice schemas
  in this folder.
- **Findings are risks, not verdicts.** Like any linter, a finding is a pattern
  worth a look, not a guaranteed bug.

It never inspects whether a field is in `required`. A present field is extracted
regardless of `required`, and under strict structured output every declared
property is emitted anyway — so **nullability (the type), not `required`, is the
lever** that lets the model report "absent". Every rule judges a field by its
type/shape (and, for two rules, its name), never its `required` membership.

---

## Running it

**CLI** (primary use; gates CI):
```bash
python -m harness.lint schemas/baseline.json      # one file
python -m harness.lint schemas/*.json             # globs
python -m harness.lint my.json \
    --sign-terms discount,markdown,chargeback \   # override the lexical vocab
    --enum-terms status,severity,region
```
Exit code: **1** if any `warn`-level finding is produced (so CI fails), **0** if
clean, **2** on usage error.

**Programmatically:**
```python
from harness.lint import lint_schema
import json

schema   = json.load(open("my_schema.json", encoding="utf-8"))
findings = lint_schema(schema)                       # -> list[Finding]
# optional: custom vocab for a non-invoice domain
findings = lint_schema(schema, sign_terms=[...], enum_terms=[...])

warns = [f for f in findings if f.severity == "warn"]
```
Each `Finding` is a dataclass: `severity` (`"warn"`/`"info"`), `path` (e.g.
`line_items[].amount`), `rule` (the id), and `message`.

---

## How it traverses the schema

`_walk_objects()` yields `(path, name, prop)` for **every** property, recursing
into:
- nested objects (`prop["properties"]`),
- array element objects (`prop["items"]["properties"]`, path suffix `[]`),
- definitions (`schema["$defs"]`).

Each yielded property is run through the per-field rules below. One rule is
schema-level (it looks at the whole schema at once) rather than per-field.

Two small helpers do the type bookkeeping:
- `_types(prop)` → the declared type(s) with `"null"` removed (so `["number","null"]` → `["number"]`).
- `_is_nullable(prop)` → `True` if the type admits `null` (`"null"`, or `"null"` inside a type list).

---

## Severity model

| Severity | Meaning | Affects exit code? |
|---|---|---|
| `warn` | A pattern measured (or strongly expected) to cause wrong output. | Yes — exit 1 |
| `info` | A weaker signal / hygiene suggestion. | No |

Findings are sorted warnings-first, then by path.

---

## The two kinds of rule

| Kind | Rules | How it decides |
|---|---|---|
| **Structural** (domain-agnostic, always on) | `no-nullable-fields`, `non-nullable-number`, `untyped-field`, `array-without-items`, `unconstrained-object`, `missing-description` | purely from the schema's shape — works on any schema |
| **Lexical** (vocabulary-driven, configurable) | `sign-convention-no-reason`, `enum-candidate` | matches the **field name** against a term list, because the relevant property (does it carry a sign? is it categorical?) cannot be read from structure alone |

The lexical rules ship with default term lists (`DEFAULT_SIGN_TERMS`,
`DEFAULT_ENUM_TERMS`) that you can replace per domain — see *Configuration*.

---

## Every rule, and exactly how it is detected

Notation: `types` = `_types(prop)`, `nullable` = `_is_nullable(prop)`,
`low` = `name.lower()`, `numeric` = `"number" in types or "integer" in types`.

### 1. `no-nullable-fields` — `warn` · structural · **measured**
**Detects:** a schema in which *no field anywhere* is nullable.
**Condition (schema-level):**
```python
not any(_is_nullable(p) for _, _, p in _walk_objects(schema))   # and the schema has properties
```
**Why:** if nothing can be `null`, the model has no legal way to say "this field
is absent", so on documents that omit an optional field it fabricates a value.
This is the headline failure in `RESULTS.md`: the all-non-nullable baseline
returned `0` for absent discount/tax; making those fields nullable lifted
absent-field accuracy **73% → 100%**.
**Fix it flags:** mark genuinely-optional fields `["<type>", "null"]`.

### 2. `non-nullable-number` — `info` · structural · **measured**
**Detects:** a numeric field that cannot be null.
**Condition (per field):**
```python
not nullable and numeric
```
**Why:** a non-nullable number is the field most likely to be fabricated as `0`
when it is absent (`0` looks like a legitimate amount, unlike an empty string).
It is `info`, not `warn`, because many numbers legitimately are non-nullable;
it is a "consider `["number","null"]` if this can be absent" nudge.
**Note:** it deliberately does **not** fire on every non-nullable field — only
numeric ones — because that is where the measured fabrication happened.

### 3. `untyped-field` — `info` · structural · hygiene
**Detects:** a property that constrains its value in no way.
**Condition:**
```python
not (prop.get("type") or "enum"/"const"/"$ref"/"anyOf"/"oneOf"/"allOf" in prop)
```
**Why:** with no type, enum, or reference, the model is unconstrained on that
field and output drifts. General schema hygiene; not specific to invoices.

### 4. `array-without-items` — `info` · structural · hygiene
**Detects:** an array property with no element schema.
**Condition:**
```python
"array" in types and not isinstance(prop.get("items"), dict)
```
**Why:** without an `items` schema the element shape is unconstrained, so the
model invents element structure. (An array of primitives, e.g. `items:
{"type":"string"}`, satisfies the rule and is *not* flagged.)

### 5. `unconstrained-object` — `info` · structural · hygiene
**Detects:** an object property with no declared `properties`.
**Condition:**
```python
"object" in types and not isinstance(prop.get("properties"), dict)
```
**Why:** a free-form object lets the model invent arbitrary keys. Declaring the
properties pins the structure.

### 6. `missing-description` — `info` · structural · **heuristic (unmeasured here)**
**Detects:** a scalar field with no `description`.
**Condition:**
```python
not prop.get("description") and not ({"object","array"} & set(types))
```
**Why:** descriptions are generally believed to help extraction, but this was
**not** measured in this folder (the agent baseline already described every
scalar field, so the rule never fired on our corpus). It is labelled advisory
precisely to stay honest about what is and isn't proven here.

### 7. `sign-convention-no-reason` — `warn` · lexical (configurable) · **measured**
**Detects:** a numeric field whose name suggests a sign-ambiguous amount, with no
reasoning prompt to state the sign.
**Condition:**
```python
numeric and any(term in low for term in sign_terms) and "X-ReasoningPrompt" not in prop
```
- `sign_terms` defaults to `("discount","adjustment","credit","refund","rebate",
  "markdown","chargeback","delta","variance")`.
- Matching is **substring** on the lowercased field name (so `discount_amount`
  and `rebate_total` both match).
**Why:** these amounts are conventionally printed negative, and the model often
returns the wrong sign. In `RESULTS.md` the baseline returned `+150`/`+48`/`+32.25`
for `-150`/`-48`/`-32.25`; adding an `X-ReasoningPrompt` stating the sign fixed
all three (present-field **88% → 94%**). The rule clears once such a field has a
reasoning prompt.
**Limit:** name-based — `concession` or a misspelling won't match unless you add
it to `sign_terms`.

### 8. `enum-candidate` — `warn` · lexical (configurable) · **measured**
**Detects:** a categorical-looking string field with no `enum`.
**Condition:**
```python
"string" in types and "enum" not in prop and low in enum_terms
```
- `enum_terms` defaults to `{status, type, currency, category, state, kind,
  priority, mode, level, method, unit, country, language, region}`.
- Matching is **exact** on the lowercased field name (so a field literally named
  `currency` matches; `customer_code` does not match `code`).
**Why:** a free-text string echoes whatever surface form the document prints; an
`enum` normalizes the value to a canonical vocabulary. In `RESULTS.md`, `currency`
printed as `Euros`/`US Dollars`/`€` scored 2–3/5 as free text and **5/5** once an
ISO-code enum was added (present-field **94% → 98%**), and its consensus
agreement rose from **87.5% → 100%**.
**Limit:** exact-name match — a categorical field named `disposition` won't match
unless you add it to `enum_terms`.

---

## Configuration

The two lexical rules are the only domain-specific part. Override their
vocabularies for your own schemas:

- **CLI:** `--sign-terms a,b,c` and `--enum-terms x,y,z` (comma-separated; replace
  the defaults entirely).
- **Python:** `lint_schema(schema, sign_terms=[...], enum_terms=[...])`.

Everything else (rules 1–6) is structural and needs no configuration.

---

## What the linter cannot do

Because detection is structural + name-based, it is fast and explainable but
**heuristic**:
- It cannot know a field's *semantics* from structure — e.g. that a string named
  `disposition` is categorical, or that `net_change` can be negative — unless the
  name is in the (configurable) term lists. A semantic/LLM checker could, but it
  would be non-deterministic, cost tokens, and need a network call; this linter
  deliberately trades that away for being free and CI-safe.
- It flags *risks*, not certainties. Confirm with the measured harness
  (`python -m harness.experiment`) when in doubt.

## How the rules map to the measurements

On this folder's four schemas the **warn count moves in lock-step with measured
accuracy**, which is the evidence that the warn-level rules encode real failures
rather than style opinions:

| Schema | Warns | Accuracy |
|---|:--:|:--:|
| baseline | 3 (no-nullable + sign-convention + enum-candidate) | 83% |
| nullable | 2 (sign-convention + enum-candidate) | 90% |
| reasoning | 1 (enum-candidate) | 95% |
| enum | 0 | 98% |

Each warning that clears corresponds to a measured accuracy gain. See
[`RESULTS.md`](RESULTS.md) for the full experiment.
