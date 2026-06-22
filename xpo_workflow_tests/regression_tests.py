#!/usr/bin/env python3
"""
XPO NEW - PRODUCTION  —  function-block regression tests (core logic).

WHAT THIS PROTECTS
------------------
The "Convert & Merge for XPO" function block contains the deterministic business
logic where "fix one customer, break another" regressions happen:

  * single-commodity customers (FilmTec / AGCO / GIMA) -> ONE aggregated commodity
    per stop, with summed weight / quantity / volume / linear-meters;
  * every other customer keeps its per-row commodities;
  * several orders for the SAME customer in one document are merged into one order;
  * MULTI-CLIENT documents (different customers) stay as separate orders and their
    sums are never mixed together;
  * customer NAME -> XPO 10-digit code resolution (via the mounted lookup table),
    incl. accent/case tolerance and the single-commodity allow-list.

HOW IT STAYS HONEST
-------------------
It does NOT contain a copy of the function code. On every run it pulls the
*current* code and the *current* customer-codes table straight from the workflow
via the `retab` CLI, then exercises the real functions. So it always tests what
is actually deployed in the draft — run it after any edit, before publishing.

It only executes the pure, deterministic helpers (grouping / merging / single-
commodity / customer lookup). It never runs the workflow end to end, so it never
calls the XPO API. Zero production side effects.

PREREQUISITES
-------------
  * Python 3.9+
  * `retab` CLI on PATH, logged in, scoped to the XPO organization
    (check with:  retab auth status   -> organization.name should be "XPO")

USAGE
-----
  python regression_tests.py                 # uses the default workflow id below
  python regression_tests.py <workflow-id>   # or pass another workflow id

Exit code 0 = all passed, 1 = at least one failure (suitable for CI / pre-publish).
"""

import sys
import json
import types
import subprocess

# Default target: XPO NEW - PRODUCTION
DEFAULT_WORKFLOW_ID = "wrk_-dF0BUWtswiL_8w3XTj6o"

# Names the function module imports from the sandbox `models` module. We stub
# them so the deployed code imports cleanly; the helpers under test never use them.
_MODEL_NAMES = [
    "Input", "Output", "XpoSpecificReference", "XpoAddress", "XpoPlanning",
    "XpoLoading", "XpoHazardGoods", "XpoDimensions", "XpoNote", "XpoCommodity",
    "XpoStop", "XpoProviderData", "XpoOutputOrder", "XpoConvertedOrder",
]


# --------------------------------------------------------------------------- #
# Loading the live function block + customer table from the workflow
# --------------------------------------------------------------------------- #
def _retab(args):
    """Run a retab CLI command and return stdout decoded as UTF-8."""
    r = subprocess.run(["retab", *args], capture_output=True)
    if r.returncode != 0:
        raise SystemExit(
            "retab %s failed:\n%s" % (" ".join(args), r.stderr.decode("utf-8", "replace"))
        )
    return r.stdout.decode("utf-8")


def _check_org():
    st = json.loads(_retab(["auth", "status", "--output", "json"]))
    org = st.get("organization", {}).get("name")
    env = st.get("environment", {}).get("name")
    if org != "XPO":
        raise SystemExit(
            "CLI is scoped to organization %r, not 'XPO'. Run:  retab org switch XPO" % org
        )
    print("Org: %s | Env: %s" % (org, env))


def load_function_block(workflow_id):
    """Return (function_code:str, table_id:str|None) for the workflow's function block."""
    blocks = json.loads(_retab(["workflows", "blocks", "list", workflow_id, "--output", "json"]))
    blocks = blocks.get("data", blocks)
    fn = next((b for b in blocks if (b.get("block_type") or b.get("type")) == "function"), None)
    if fn is None:
        raise SystemExit("No function block found in workflow %s" % workflow_id)
    full = json.loads(_retab(["workflows", "blocks", "get", workflow_id, fn["id"], "--output", "json"]))
    full = full.get("data", full)
    cfg = full["config"]
    tables = (cfg.get("mounts") or {}).get("tables") or []
    table_id = tables[0]["table_id"] if tables else None
    return cfg["code"], table_id


def load_customer_table(table_id):
    """Return {customer_name: code} and {code: customer_name} from the live table."""
    csv_text = _retab(["tables", "download", table_id])
    name_to_code, code_to_name = {}, {}
    lines = [ln for ln in csv_text.splitlines() if ln.strip()]
    for ln in lines[1:]:  # skip header
        parts = ln.split(",")
        if len(parts) < 2:
            continue
        name, code = parts[0].strip().strip('"'), parts[1].strip().strip('"')
        if name and code:
            name_to_code[name] = code
            code_to_name[code] = name
    return name_to_code, code_to_name


def load_live_namespace(function_code, table_id):
    """Exec the deployed function code with a stubbed `models` module and return its
    namespace, with customer-code resolution pointed at a freshly downloaded copy
    of the live table."""
    # Stub the sandbox `models` module so `from models import ...` works.
    stub = types.ModuleType("models")

    class _Stub:
        def __init__(self, **kw):
            self.__dict__.update(kw)

        def model_dump(self, *a, **k):
            return dict(self.__dict__)

    for name in _MODEL_NAMES:
        setattr(stub, name, _Stub)
    sys.modules["models"] = stub

    ns = {}
    exec(compile(function_code, "<xpo_function_block>", "exec"), ns)

    # Point the customer-code lookup at a local copy of the live table and reset
    # its cache, so _resolve_xpo_customer_code reads real, current data.
    if table_id and "_load_xpo_customer_codes" in ns:
        csv_text = _retab(["tables", "download", table_id])
        path = "_live_xpo_customer_codes.csv"
        with open(path, "w", encoding="utf-8") as fh:
            fh.write(csv_text)
        ns["XPO_CUSTOMER_CODES_CSV_PATH"] = path
        ns["_XPO_CUSTOMER_CODES_CACHE"] = None
    return ns


# --------------------------------------------------------------------------- #
# Test data builders
# --------------------------------------------------------------------------- #
def commodity(weight, qty, volume, lm, desc="OTHER", code="OTHER", packaging="PAL"):
    return {
        "weight": weight, "code": code, "description": desc,
        "packagingCode": packaging, "packagingQuantity": qty,
        "volume": volume, "linearMeters": lm, "hazardGoods": {},
    }


def order_dict(customer_code, commodities, ship_ref="REF", provider_id="pid"):
    """Build one converted-order dict in the exact shape _build_payloads consumes
    (pickup + delivery stop, identical commodities, mirroring transform())."""
    def stop(stype, seq):
        return {
            "stopType": stype, "stopSequence": seq,
            "address": {"addressName": "Site"},
            "specificReferences": [{"code": "CG", "value": ship_ref}],
            "commodities": [dict(c) for c in commodities],
            "planning": {}, "loading": {},
        }
    return {
        "providerData": {"name": "Cube AI", "operationCode": "C", "sourceReference": "TEMP"},
        "order": {
            "providerOrderID": provider_id, "shipmentReference": ship_ref,
            "specificReferences": [{"code": "CG", "value": ship_ref}],
            "stops": [stop("P", 1), stop("D", 2)],
            "agency": {"reportingCode": "f877"},
            "customer": {"XPOcustomerCode": customer_code},
            "pricing": {"taxableQuantity": 0},
        },
    }


def counts(payload):
    """Commodity count per stop, e.g. [1, 1]."""
    return [len(s["commodities"]) for s in payload["order"]["stops"]]


def first_comm(payload, stop_index=0):
    return payload["order"]["stops"][stop_index]["commodities"][0]


# --------------------------------------------------------------------------- #
# Assertions
# --------------------------------------------------------------------------- #
_FAILURES = []


def check(name, cond, detail=""):
    if cond:
        print("  [PASS] %s" % name)
    else:
        print("  [FAIL] %s%s" % (name, (" -- " + detail) if detail else ""))
        _FAILURES.append(name)


# --------------------------------------------------------------------------- #
# The tests
# --------------------------------------------------------------------------- #
def run_tests(ns, name_to_code, code_to_name):
    build = ns["_build_payloads"]
    resolve = ns["_resolve_xpo_customer_code"]
    single = ns["SINGLE_COMMODITY_CUSTOMERS"]

    # Resolve the customers we reference, from the live table (fail loudly if missing).
    def code_for(display_name):
        c = resolve(display_name)
        if not c:
            raise SystemExit("Customer %r not found in live table; update the test or the table." % display_name)
        return c

    FILMTEC = code_for("FilmTec")
    AGCO = code_for("AGCO Beauvais")
    GIMA = code_for("GIMA Beauvais")
    # A customer that must NOT be single-commodity (any non-FilmTec/AGCO/GIMA row).
    other_name = next(
        (n for n, c in name_to_code.items() if c not in single), None
    )
    if other_name is None:
        raise SystemExit("Every customer is in SINGLE_COMMODITY_CUSTOMERS? Check the table/allow-list.")
    OTHER = name_to_code[other_name]

    # Real production commodity values from the FilmTec bug (run_jrxEjv832...): 2 rows.
    REAL_FILMTEC = [
        commodity(2924, 4, 5.718, 1.43),
        commodity(731, 1, 1.43, 1.43),
    ]
    SUM = {"weight": 3655, "qty": 5, "volume": 7.148, "lm": 2.86}

    print("\n# Allow-list sanity")
    check("FilmTec, AGCO, GIMA are all in SINGLE_COMMODITY_CUSTOMERS",
          {FILMTEC, AGCO, GIMA} <= single, "set=%s" % sorted(single))
    check("a non-listed customer (%s) is NOT single-commodity" % other_name,
          OTHER not in single)
    for c in sorted(single):
        check("allow-list code %s maps to a real customer (%s)" % (c, code_to_name.get(c, "??")),
              c in code_to_name, "code not in customer table -- typo?")

    print("\n# Single-commodity customers collapse a multi-row order to ONE summed commodity")
    for label, cc in [("FilmTec", FILMTEC), ("AGCO", AGCO), ("GIMA", GIMA)]:
        out = build([order_dict(cc, REAL_FILMTEC)])
        check("%s: exactly one output order" % label, len(out) == 1)
        check("%s: 1 commodity per stop [1,1]" % label, counts(out[0]) == [1, 1], "got %s" % counts(out[0]))
        c = first_comm(out[0])
        check("%s: weight summed (=%d)" % (label, SUM["weight"]), c["weight"] == SUM["weight"], "got %s" % c["weight"])
        check("%s: quantity summed (=%d)" % (label, SUM["qty"]), c["packagingQuantity"] == SUM["qty"], "got %s" % c["packagingQuantity"])
        check("%s: volume summed (=%s)" % (label, SUM["volume"]), abs(c["volume"] - SUM["volume"]) < 1e-9, "got %s" % c["volume"])
        check("%s: linear-meters summed (=%s)" % (label, SUM["lm"]), abs(c["linearMeters"] - SUM["lm"]) < 1e-9, "got %s" % c["linearMeters"])
        check("%s: delivery stop also summed" % label, first_comm(out[0], 1)["weight"] == SUM["weight"])

    print("\n# Non-listed customers KEEP their per-row commodities")
    out = build([order_dict(OTHER, REAL_FILMTEC)])
    check("%s: still 2 commodities per stop [2,2]" % other_name, counts(out[0]) == [2, 2], "got %s" % counts(out[0]))

    print("\n# Single-commodity customer with a single row stays single (idempotent, no crash)")
    out = build([order_dict(FILMTEC, [REAL_FILMTEC[0]])])
    check("FilmTec single-row stays [1,1]", counts(out[0]) == [1, 1], "got %s" % counts(out[0]))
    check("FilmTec single-row weight untouched", first_comm(out[0])["weight"] == 2924)

    print("\n# Several orders for the SAME customer in one document merge into ONE order")
    out = build([order_dict(OTHER, [REAL_FILMTEC[0]], ship_ref="A", provider_id="a"),
                 order_dict(OTHER, [REAL_FILMTEC[1]], ship_ref="B", provider_id="b")])
    check("%s: two same-customer orders -> 1 merged order" % other_name, len(out) == 1, "got %d" % len(out))
    check("%s: merged order aggregates to 1 commodity" % other_name, counts(out[0]) == [1, 1], "got %s" % counts(out[0]))
    check("%s: merged weight = 2924+731 = 3655" % other_name, first_comm(out[0])["weight"] == 3655,
          "got %s" % first_comm(out[0])["weight"])

    print("\n# MULTI-CLIENT: different customers stay separate and sums are NOT mixed")
    out = build([order_dict(FILMTEC, REAL_FILMTEC, ship_ref="F", provider_id="f"),
                 order_dict(OTHER, REAL_FILMTEC, ship_ref="O", provider_id="o")])
    check("multiclient -> 2 separate output orders", len(out) == 2, "got %d" % len(out))
    by_code = {p["order"]["customer"]["XPOcustomerCode"]: p for p in out}
    check("FilmTec part collapsed to [1,1]", counts(by_code[FILMTEC]) == [1, 1], "got %s" % counts(by_code[FILMTEC]))
    check("%s part keeps [2,2]" % other_name, counts(by_code[OTHER]) == [2, 2], "got %s" % counts(by_code[OTHER]))
    check("FilmTec sum is only its own (3655, not mixed)", first_comm(by_code[FILMTEC])["weight"] == 3655,
          "got %s" % first_comm(by_code[FILMTEC])["weight"])

    print("\n# Customer NAME -> XPO code resolution (live table)")
    check("FilmTec -> %s" % FILMTEC, resolve("FilmTec") == FILMTEC)
    check("AGCO Beauvais -> %s" % AGCO, resolve("AGCO Beauvais") == AGCO)
    check("GIMA Beauvais -> %s" % GIMA, resolve("GIMA Beauvais") == GIMA)
    check("accent/case tolerant: '  filmtec ' -> %s" % FILMTEC, resolve("  filmtec ") == FILMTEC)
    check("unknown customer name -> '' (so transform raises a clear error)", resolve("Totally Unknown SARL") == "")


def main():
    workflow_id = sys.argv[1] if len(sys.argv) > 1 else DEFAULT_WORKFLOW_ID
    print("XPO function-block regression tests")
    print("workflow: %s" % workflow_id)
    _check_org()
    code, table_id = load_function_block(workflow_id)
    print("loaded live function code: %d lines | customer table: %s" % (len(code.splitlines()), table_id))
    name_to_code, code_to_name = load_customer_table(table_id) if table_id else ({}, {})
    ns = load_live_namespace(code, table_id)
    run_tests(ns, name_to_code, code_to_name)

    print("\n" + "=" * 60)
    if _FAILURES:
        print("RESULT: %d FAILURE(S): %s" % (len(_FAILURES), _FAILURES))
        sys.exit(1)
    print("RESULT: ALL TESTS PASSED")
    sys.exit(0)


if __name__ == "__main__":
    main()
