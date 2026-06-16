"""Generate the invoice test corpus for the schema-quality cookbook.

All five documents are the SAME document type (an invoice / facture) and are
meant to be processed by ONE schema (one extract block). They differ only in
*form* — which optional fields are present:

    field              full  no_name  no_code  minimal  mixed
    customer_name       x       .        x        .       x
    customer_code       x       x        .        .       x
    purchase_order      x       x        .        .       .
    discount            x       x        x        .       .
    due_date            x       x        x        .       x
    tax                 x       x        x        .       x

(`x` = present, `.` = absent)

To exercise more than one failure mode, the corpus also varies *convention*,
not just presence:
  - `discount`, when present, is printed as a negative number (sign convention).
  - `invoice_date` is printed in European DD/MM/YYYY on several invoices
    (e.g. "07/02/2024" = 7 February), which a day/month-agnostic reader can
    misinterpret. Ground truth is always ISO 8601.

This is the right shape for testing nullable types: the same field is present
on some invoices and genuinely absent on others. A schema that marks such a
field `required` (the default a generator emits) forces the model to invent a
value on the invoices where it is missing.

The values are fixed (deterministic) so re-running produces byte-stable PDFs,
and documents/manifest.json records the correct value of every field.

Run:  python generate_documents.py        (requires: pip install reportlab)
"""

from __future__ import annotations

import json
import os

from reportlab.lib import colors
from reportlab.lib.pagesizes import LETTER
from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet
from reportlab.lib.units import inch
from reportlab.platypus import (
    Paragraph,
    SimpleDocTemplate,
    Spacer,
    Table,
    TableStyle,
)

HERE = os.path.dirname(os.path.abspath(__file__))
OUT_DIR = os.path.join(HERE, "documents")

styles = getSampleStyleSheet()
H1 = ParagraphStyle("H1", parent=styles["Heading1"], fontSize=18, spaceAfter=4)
H2 = ParagraphStyle("H2", parent=styles["Heading2"], fontSize=11, spaceAfter=3)
BODY = ParagraphStyle("BODY", parent=styles["BodyText"], fontSize=9, leading=12)
SMALL = ParagraphStyle("SMALL", parent=styles["BodyText"], fontSize=8.5, leading=11)


def _doc(filename):
    return SimpleDocTemplate(
        os.path.join(OUT_DIR, filename), pagesize=LETTER,
        leftMargin=0.8 * inch, rightMargin=0.8 * inch,
        topMargin=0.8 * inch, bottomMargin=0.8 * inch, title=filename,
    )


def _kv_table(rows):
    t = Table(rows, colWidths=[1.4 * inch, 2.1 * inch])
    t.setStyle(TableStyle([
        ("FONTSIZE", (0, 0), (-1, -1), 8.5),
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        ("FONTNAME", (0, 0), (0, -1), "Helvetica-Bold"),
        ("LEFTPADDING", (0, 0), (-1, -1), 0),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 2),
    ]))
    return t


# Fixed vendor for every invoice (held constant so it is not a variable).
VENDOR = {
    "name": "Northwind Trading Co.",
    "address": "12 Harbour Road, Bristol BS1 5TT, United Kingdom",
    "vat": "GB 814 2299 07",
}

# The five invoices. Optional fields set to None are omitted from the PDF.
INVOICES = [
    {
        "file": "invoice_full.pdf",
        "invoice_number": "INV-1001",
        "invoice_date": "2024-01-15",
        "due_date": "2024-02-14",
        "currency": "EUR",
        "customer_company": "Lumiere Studios SARL",
        "customer_name": "Claire Fontaine",
        "customer_code": "CUST-4471",
        "purchase_order": "PO-22910",
        "line_items": [
            ("Design consulting (hrs)", 10, 120.00, 1200.00),
            ("Stock photo license", 1, 300.00, 300.00),
        ],
        "subtotal": 1500.00, "discount": -150.00, "tax": 270.00, "total": 1620.00,
    },
    {
        "file": "invoice_no_name.pdf",
        "invoice_number": "INV-1002",
        "invoice_date": "2024-02-07",
        "invoice_date_printed": "07/02/2024",
        "due_date": "2024-02-21",
        "currency": "EUR",
        "customer_company": "Atelier Bois & Co",
        "customer_name": None,
        "customer_code": "CUST-3320",
        "purchase_order": "PO-22955",
        "line_items": [
            ("Oak panels (m2)", 20, 45.00, 900.00),
            ("Delivery", 1, 60.00, 60.00),
        ],
        "subtotal": 960.00, "discount": -48.00, "tax": 182.40, "total": 1094.40,
    },
    {
        "file": "invoice_no_code.pdf",
        "invoice_number": "INV-1003",
        "invoice_date": "2024-02-03",
        "invoice_date_printed": "03/02/2024",
        "due_date": "2024-03-04",
        "currency": "EUR",
        "customer_company": "Cafe Central GmbH",
        "customer_name": "Markus Weber",
        "customer_code": None,
        "purchase_order": None,
        "line_items": [
            ("Espresso beans (kg)", 30, 18.50, 555.00),
            ("Filter packs", 12, 7.50, 90.00),
        ],
        "subtotal": 645.00, "discount": -32.25, "tax": 122.55, "total": 735.30,
    },
    {
        "file": "invoice_minimal.pdf",
        "invoice_number": "INV-1004",
        "invoice_date": "2024-02-10",
        "invoice_date_printed": "10/02/2024",
        "due_date": None,
        "currency": "EUR",
        "customer_company": "Nordic Print AB",
        "customer_name": None,
        "customer_code": None,
        "purchase_order": None,
        "line_items": [
            ("Poster printing A1", 100, 4.20, 420.00),
        ],
        "subtotal": 420.00, "discount": None, "tax": None, "total": 420.00,
    },
    {
        "file": "invoice_mixed.pdf",
        "invoice_number": "INV-1005",
        "invoice_date": "2024-03-05",
        "invoice_date_printed": "05/03/2024",
        "due_date": "2024-04-04",
        "currency": "EUR",
        "customer_company": "Verde Landscaping SL",
        "customer_name": "Lucia Ramirez",
        "customer_code": "CUST-7782",
        "purchase_order": None,
        "line_items": [
            ("Garden maintenance (hrs)", 25, 35.00, 875.00),
            ("Plant supply", 1, 140.00, 140.00),
        ],
        "subtotal": 1015.00, "discount": None, "tax": 203.00, "total": 1218.00,
    },
]


def build_invoice(inv):
    s = []
    s.append(Paragraph("INVOICE", H1))
    s.append(Paragraph(VENDOR["name"], H2))
    s.append(Paragraph(f"{VENDOR['address']}<br/>VAT: {VENDOR['vat']}", SMALL))
    s.append(Spacer(1, 12))

    # Header key/values; optional lines appear only when present.
    meta = [["Invoice number", inv["invoice_number"]],
            ["Invoice date", inv.get("invoice_date_printed", inv["invoice_date"])]]
    if inv["due_date"]:
        meta.append(["Date due", inv["due_date"]])
    meta.append(["Currency", inv["currency"]])
    s.append(_kv_table(meta))
    s.append(Spacer(1, 12))

    # Bill-to block.
    s.append(Paragraph("Bill to", H2))
    bill = inv["customer_company"]
    if inv["customer_name"]:
        bill = f"{inv['customer_name']}<br/>{bill}"
    if inv["customer_code"]:
        bill += f"<br/>Customer code: {inv['customer_code']}"
    if inv["purchase_order"]:
        bill += f"<br/>Purchase order: {inv['purchase_order']}"
    s.append(Paragraph(bill, SMALL))
    s.append(Spacer(1, 16))

    # Line items.
    rows = [["Description", "Qty", "Unit price", "Amount"]]
    for d, q, up, amt in inv["line_items"]:
        rows.append([d, str(q), f"{up:,.2f}", f"{amt:,.2f}"])
    t = Table(rows, colWidths=[3.6 * inch, 0.7 * inch, 1.1 * inch, 1.1 * inch])
    t.setStyle(TableStyle([
        ("FONTSIZE", (0, 0), (-1, -1), 9),
        ("GRID", (0, 0), (-1, -1), 0.4, colors.HexColor("#bbbbbb")),
        ("BACKGROUND", (0, 0), (-1, 0), colors.HexColor("#222222")),
        ("TEXTCOLOR", (0, 0), (-1, 0), colors.white),
        ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
        ("ALIGN", (1, 0), (-1, -1), "RIGHT"),
    ]))
    s.append(t)
    s.append(Spacer(1, 12))

    # Totals; discount line appears only when present.
    totals = [["Subtotal", f"{inv['subtotal']:,.2f}"]]
    if inv["discount"] is not None:
        totals.append(["Discount", f"{inv['discount']:,.2f}"])
    if inv["tax"] is not None:
        totals.append(["VAT (20%)", f"{inv['tax']:,.2f}"])
    totals.append(["Total", f"{inv['total']:,.2f}"])
    tt = Table(totals, colWidths=[4.3 * inch, 1.2 * inch])
    tt.setStyle(TableStyle([
        ("FONTSIZE", (0, 0), (-1, -1), 9),
        ("ALIGN", (1, 0), (1, -1), "RIGHT"),
        ("FONTNAME", (0, -1), (-1, -1), "Helvetica-Bold"),
        ("LINEABOVE", (0, -1), (-1, -1), 0.6, colors.black),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 3),
    ]))
    s.append(tt)
    s.append(Spacer(1, 18))
    s.append(Paragraph("Thank you for your business.", SMALL))
    _doc(inv["file"]).build(s)


def ground_truth(inv):
    """Canonical correct values. None means the field is absent on this invoice."""
    return {
        "invoice_number": inv["invoice_number"],
        "invoice_date": inv["invoice_date"],
        "due_date": inv["due_date"],
        "currency": inv["currency"],
        "customer_company": inv["customer_company"],
        "customer_name": inv["customer_name"],
        "customer_code": inv["customer_code"],
        "purchase_order": inv["purchase_order"],
        "subtotal": inv["subtotal"],
        "discount": inv["discount"],
        "tax": inv["tax"],
        "total": inv["total"],
    }


def write_manifest():
    docs = []
    for inv in INVOICES:
        gt = ground_truth(inv)
        docs.append({
            "file": inv["file"],
            "present": [k for k, v in gt.items() if v is not None],
            "absent": [k for k, v in gt.items() if v is None],
            "ground_truth": gt,
        })
    manifest = {
        "description": (
            "Invoice corpus for the schema-quality cookbook. All documents are "
            "invoices processed by one schema; they differ only in which "
            "optional fields are present. ground_truth holds the correct value "
            "of each field (null = absent on that invoice). Field names are "
            "canonical; schemas/field_map.json maps them to each schema's paths."
        ),
        "documents": docs,
    }
    with open(os.path.join(OUT_DIR, "manifest.json"), "w", encoding="utf-8") as f:
        json.dump(manifest, f, indent=2, ensure_ascii=False)


def main():
    os.makedirs(OUT_DIR, exist_ok=True)
    for inv in INVOICES:
        build_invoice(inv)
    write_manifest()
    print(f"Wrote {len(INVOICES)} invoice PDFs + manifest.json to {OUT_DIR}")


if __name__ == "__main__":
    main()
