# Classify

Endpoint: `POST /v1/documents/classify`

## Use it when

- You need one label from a fixed set of categories
- You want a lightweight routing step before more expensive processing
- The result can be decided from document semantics rather than exact field extraction

## Minimal Python

```python
from retab import Retab

client = Retab()
result = client.documents.classify(
    document="document.pdf",
    model="retab-small",
    categories=[
        {"name": "invoice", "description": "Invoice documents"},
        {"name": "receipt", "description": "Receipt documents"},
        {"name": "contract", "description": "Contract documents"},
    ],
)

print(result.result.classification)
print(result.result.reasoning)
```

## Minimal Node

```ts
import { Retab } from "@retab/node";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });

const result = await client.documents.classify({
  document: "document.pdf",
  model: "retab-small",
  categories: [
    { name: "invoice", description: "Invoice documents" },
    { name: "receipt", description: "Receipt documents" },
    { name: "contract", description: "Contract documents" },
  ],
});

console.log(result.result.classification);
console.log(result.result.reasoning);
```

## Minimal REST

```bash
curl -X POST "https://api.retab.com/v1/documents/classify" \
  -H "Api-Key: $RETAB_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "filename": "document.pdf",
      "url": "data:application/pdf;base64,..."
    },
    "model": "retab-small",
    "categories": [
      { "name": "invoice", "description": "Invoice documents" },
      { "name": "receipt", "description": "Receipt documents" },
      { "name": "contract", "description": "Contract documents" }
    ]
  }'
```

## Useful fields

- `categories`: required; each item needs `name` and `description`
- `first_n_pages`: limit analysis when early pages are enough
- `context`: add business context for ambiguous categories

## Response shape

- `result.classification`
- `result.reasoning`

## Guidance

- Keep categories mutually exclusive when possible.
- Use `first_n_pages` when the label can be determined early and latency matters.
- Use `classify` for routing. Use `extract` when you need structured values.
