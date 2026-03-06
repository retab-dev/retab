# Split

Endpoint: `POST /v1/documents/split`

## Use it when

- One uploaded file contains multiple subdocuments
- You need page ranges for each document type
- You need partitions inside one type, such as one invoice number per section

## Minimal Python

```python
from retab import Retab

client = Retab()
result = client.documents.split(
    document="batch.pdf",
    model="retab-small",
    subdocuments=[
        {"name": "invoice", "description": "Invoice documents"},
        {"name": "receipt", "description": "Receipt documents"},
    ],
)

print(result.splits)
```

## Minimal Node

```ts
import { Retab } from "@retab/node";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });

const result = await client.documents.split({
  document: "batch.pdf",
  model: "retab-small",
  subdocuments: [
    { name: "invoice", description: "Invoice documents" },
    { name: "receipt", description: "Receipt documents" },
  ],
});

console.log(result.splits);
```

## Minimal REST

```bash
curl -X POST "https://api.retab.com/v1/documents/split" \
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

- `splits[].name`
- `splits[].pages`
- `splits[].partitions[]` when `partition_key` is used

## Guidance

- Write distinct subdocument descriptions. Overlapping labels make routing worse.
- Add `partition_key` when one subdocument type repeats inside the same file.
- Use `split` before `extract` when a bundle must be separated first.
