# Parse

Endpoint: `POST /v1/parses`

## Use it when

- You need readable text for prompting, indexing, or debugging
- You want page-by-page output instead of schema extraction
- You need table output in `html`, `markdown`, `yaml`, or `json`
- You want free text, not structured field extraction
- You want the result persisted and retrievable later via `client.parses.get(id)` / `client.parses.list()`

## Minimal Python

```python
from retab import Retab

client = Retab()
parse = client.parses.create(
    document="document.pdf",
    model="retab-small",
    table_parsing_format="markdown",
    image_resolution_dpi=192,
)

print(parse.id)
print(parse.output.pages)
print(parse.output.text)
print(parse.file.filename)
```

## Minimal Node

```ts
import { Retab } from "@retab/node";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });

const parse = await client.parses.create({
  document: "document.pdf",
  model: "retab-small",
  table_parsing_format: "markdown",
  image_resolution_dpi: 192,
});

console.log(parse.id);
console.log(parse.output.pages);
console.log(parse.output.text);
console.log(parse.file.filename);
```

## Minimal REST

```bash
curl -X POST "https://api.retab.com/v1/parses" \
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

- `id`: parse resource id
- `file.id`, `file.filename`, `file.mime_type`
- `output.pages`: page-by-page parsed content
- `output.text`: full document content concatenated
- `model`, `table_parsing_format`, `image_resolution_dpi`
- `usage.page_count`, `usage.credits`
- `created_at`, `updated_at`

## Listing and retrieving parses

```python
parse = client.parses.get("parse_01G34H8J2K")
for item in client.parses.list(limit=20).data:
    ...
client.parses.delete(parse.id)
```

## Guidance

- Use `parses.create` when the task only needs free text.
- Use `extract` directly when you need schema-shaped output.
- Use `table_parsing_format` when downstream code needs predictable table output.
- Start at `192` DPI, lower to `96` for speed, and raise toward `300` for hard OCR cases.
- Do not rely on unsupported parse fields like `browser_canvas`.
