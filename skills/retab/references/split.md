# Split

Endpoint: `POST /v1/splits`

## Use it when

- One uploaded file contains multiple subdocuments
- You need page assignments for each subdocument type
- The same subdocument can appear multiple times in one file
- You want to route each detected subdocument into a different downstream step

Do not use `split` for key-based grouping inside one homogeneous document type. Use `POST /v1/partitions` for that.

## Minimal Python

```python
from retab import Retab

client = Retab()
result = client.splits.create(
    document="batch.pdf",
    model="retab-small",
    subdocuments=[
        {"name": "invoice", "description": "Invoice documents"},
        {"name": "receipt", "description": "Receipt documents"},
    ],
)

print(result.output)
```

## Minimal Node

```ts
import { Retab } from "@retab/node";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });

const result = await client.splits.create({
  document: "batch.pdf",
  model: "retab-small",
  subdocuments: [
    { name: "invoice", description: "Invoice documents" },
    { name: "receipt", description: "Receipt documents" },
  ],
});

console.log(result.output);
```

## Minimal REST

```bash
curl -X POST "https://api.retab.com/v1/splits" \
  -H "Api-Key: $RETAB_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "filename": "batch.pdf",
      "url": "data:application/pdf;base64,..."
    },
    "model": "retab-small",
    "subdocuments": [
      { "name": "invoice", "description": "Invoice documents" },
      { "name": "receipt", "description": "Receipt documents" }
    ]
  }'
```

## Useful fields

- `subdocuments`: required; each item needs `name`, optional `description`, and optional `allow_multiple_instances`
- `instructions`: optional free-form instructions to steer the split
- `n_consensus`: raise it when split boundaries are business-critical
- `bust_cache`: force a fresh run

## Response shape

- `output[].name`
- `output[].pages`
- `consensus.likelihoods[]` with `name` and `pages` when `n_consensus > 1`
- `consensus.choices[]` as raw voter outputs when `n_consensus > 1`

## Guidance

- Write distinct subdocument descriptions. Overlapping labels make routing worse.
- Use `allow_multiple_instances` when one subdocument type can repeat.
- Use `split` before `extract` when a bundle must be separated first.
- Use `partition` after `split` if a specific split output then needs key-based grouping.
