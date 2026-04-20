# Extract

Endpoint: `POST /v1/extractions`

## Use it when

- You need structured JSON back from a document
- The target shape is known or can be written as JSON Schema
- You want a persisted extraction resource you can fetch, list, delete, or source later

## Minimal Python

```python
from retab import Retab

client = Retab()
response = client.extractions.create(
    document="invoice.pdf",
    model="retab-small",
    json_schema={
        "type": "object",
        "properties": {
            "invoice_number": {"type": "string"},
            "total_amount": {"type": "number"}
        },
        "required": ["invoice_number", "total_amount"]
    },
)

print(response.id)
print(response.output)
```

## Minimal Node

```ts
import { Retab } from "@retab/node";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });

const response = await client.extractions.create({
  document: "invoice.pdf",
  model: "retab-small",
  json_schema: {
    type: "object",
    properties: {
      invoice_number: { type: "string" },
      total_amount: { type: "number" },
    },
    required: ["invoice_number", "total_amount"],
  },
});

console.log(response.id);
console.log(response.output);
```

## Minimal REST

```bash
curl -X POST "https://api.retab.com/v1/extractions" \
  -H "Api-Key: $RETAB_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "filename": "invoice.pdf",
      "url": "data:application/pdf;base64,..."
    },
    "model": "retab-small",
    "json_schema": {
      "type": "object",
      "properties": {
        "invoice_number": { "type": "string" },
        "total_amount": { "type": "number" }
      },
      "required": ["invoice_number", "total_amount"]
    }
  }'
```

## Useful fields

- `json_schema`: required
- `n_consensus`: defaults to `1`; raise it when accuracy matters more than speed
- `image_resolution_dpi`: defaults to `192`
- `instructions`: optional free-form instructions to steer the extraction (domain hints, tone, iteration context)
- `additional_messages`: append raw chat messages to the LLM request when you need full message-level control

## Consensus

Retab extraction is built on top of LLM calls, and LLM outputs are not perfectly deterministic. `n_consensus` tells Retab to run the same extraction multiple times and reconcile the results into one final JSON object.

- `n_consensus=1`: one extraction pass, cheapest and fastest
- `n_consensus>1`: multiple extraction passes, then field-by-field reconciliation

In practice, consensus does two things:

- It improves robustness when fields are ambiguous, noisy, or hard to read
- It produces more useful `likelihoods`, because agreement across runs becomes a signal of confidence

Retab reconciles values by type:

- Scalars such as booleans and exact strings lean on voting or matching
- Numbers are grouped by closeness rather than exact equality
- Nested objects are reconciled field by field
- Arrays are reconciled element by element after alignment

### When to increase it

Raise `n_consensus` when:

- The extraction is business-critical
- The document quality is poor
- The schema contains ambiguous fields
- You plan to gate downstream automation on confidence

Keep `n_consensus=1` when:

- You need low latency
- The fields are simple and stable
- You are still iterating on the schema and want quick feedback

### Tradeoff

Consensus is not free. More passes mean more latency and more cost, so use it selectively instead of turning it on everywhere.

### How to read likelihoods

`consensus.likelihoods` mirrors the extracted structure and reflects how strongly the runs agreed on each field.

- High likelihood: the runs converged on the same answer
- Lower likelihood: the field was ambiguous, noisy, or interpreted in multiple ways

If a field keeps getting low likelihoods, first improve the schema before just increasing `n_consensus`. Better field names, tighter descriptions, and stricter types usually help more than extra retries.

## Response shape

- `id`: extraction resource id
- `output`: extracted JSON
- `consensus.likelihoods`: per-field confidence values
- `usage`: token and cost-related metadata

## Guidance

- If no schema exists yet, draft one before implementing `extract`.
- Keep schemas small and explicit. Overly broad schemas reduce reliability.
- Use `extract` only when the output must be structured. Otherwise use `parse`.
