# Common

## Install and authenticate

- Python: `pip install retab`
- Node: `npm install @retab/node`
- REST base URL: `https://api.retab.com`
- Auth header: `Api-Key: $RETAB_API_KEY`

Python:

```python
import os

from retab import Retab

client = Retab(api_key=os.environ["RETAB_API_KEY"])
```

Node:

```ts
import { Retab } from "@retab/node";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });
```

## Shared request conventions

- Default to `model="retab-small"` unless the user asks for a different tradeoff.
- For REST, send JSON with `document: { "filename": "...", "url": "data:...base64,..." }`.
- For SDKs, prefer a local path string when the file is already on disk.
- Keep uploads focused. Trim or split overly large documents before sending them.
- Use generous timeouts for slow or multi-page documents.

## Choose the route

- `parse`: convert a document into page-by-page text or structured table output
- `extract`: map a document into a JSON schema
- `split`: assign page ranges to named subdocuments
- `edit`: fill or update a form-like document from natural-language instructions
- `classify`: choose one category from a fixed list

## REST skeleton

```bash
curl -X POST "https://api.retab.com/v1/documents/ROUTE" \
  -H "Api-Key: $RETAB_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "filename": "document.pdf",
      "url": "data:application/pdf;base64,..."
    },
    "model": "retab-small"
  }'
```
