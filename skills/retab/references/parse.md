# Parse

Endpoint: `POST /v1/documents/parse`

## Use it when

- You need readable text for prompting, indexing, or debugging
- You want page-by-page output instead of schema extraction
- You need table output in `html`, `markdown`, `yaml`, or `json`
- You want free text, not structured field extraction

## Minimal Python

```python
from retab import Retab

client = Retab()
result = client.documents.parse(
    document="document.pdf",
    model="retab-small",
    table_parsing_format="markdown",
    image_resolution_dpi=192,
)

print(result.pages)
print(result.text)
print(result.document.filename)
```

## Minimal Node

```ts
import { Retab } from "@retab/node";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });

const result = await client.documents.parse({
  document: "document.pdf",
  model: "retab-small",
  table_parsing_format: "markdown",
  image_resolution_dpi: 192,
});

console.log(result.pages);
console.log(result.text);
console.log(result.document.filename);
```

## Minimal REST

```bash
curl -X POST "https://api.retab.com/v1/documents/parse" \
  -H "Api-Key: $RETAB_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "filename": "document.pdf",
      "url": "data:application/pdf;base64,..."
    },
    "model": "retab-small",
    "table_parsing_format": "markdown",
    "image_resolution_dpi": 192
  }'
```

## Useful fields

- `table_parsing_format`: defaults to `html`
- `image_resolution_dpi`: defaults to `192`
- Parse is non-streaming in the SDK
- The documented SDK surface is `document`, `model`, `table_parsing_format`, `image_resolution_dpi`

## Response shape

- `document.id`
- `document.filename`
- `document.mime_type`
- `pages`: page-by-page parsed content
- `text`: full document content concatenated
- `usage.page_count`
- `usage.credits`

## Guidance

- Use `parse` when the task only needs free text.
- Use `extract` directly when you need schema-shaped output.
- Use `table_parsing_format` when downstream code needs predictable table output.
- Start at `192` DPI, lower to `96` for speed, and raise toward `300` for hard OCR cases.
- Do not rely on unsupported parse fields like `browser_canvas`.
