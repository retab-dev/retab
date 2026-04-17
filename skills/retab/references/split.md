# Split

Endpoint: `POST /v1/splits`

## Use it when

- One uploaded file contains multiple subdocuments
- You need the assigned pages for each document type
- You need partitions inside one type, such as one invoice number per section

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

- `subdocuments`: required; each item needs `name` and `description`
- `partition_key`: optional per subdocument; use it to break one label into repeated items
- `context`: add domain or batch context
- `n_consensus`: raise it when split boundaries are business-critical

## Response shape

- `output[].name`
- `output[].pages`
- `output[].partitions[]` when `partition_key` is used
- `consensus.likelihoods[]` when `n_consensus > 1`
- `consensus.choices[]` as raw voter outputs when `n_consensus > 1`

## Guidance

- Write distinct subdocument descriptions. Overlapping labels make routing worse.
- Add `partition_key` when one subdocument type repeats inside the same file.
- Use `split` before `extract` when a bundle must be separated first.
