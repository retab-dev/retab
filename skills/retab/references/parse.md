# Parse

Endpoint: `POST /v1/documents/parse`

## Use it when

- You need readable text for prompting, indexing, or debugging
- You want page-by-page output instead of schema extraction
- You need table output in `html`, `markdown`, `yaml`, or `json`

## Minimal Python

```python
from retab import Retab

client = Retab()
result = client.documents.parse(
    document="document.pdf",
    model="retab-small",
)

print(result.pages)
print(result.text)
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
    "model": "retab-small"
  }'
```

## Useful fields

- `table_parsing_format`: defaults to `html`
- `image_resolution_dpi`: defaults to `192`

## Response shape

- `pages`: page-by-page parsed content
- `text`: full document content concatenated
- `usage.page_count`
- `usage.credits`

## Guidance

- Use `parse` before `extract` when the task only needs free text.
- Use a lower DPI only when speed matters more than OCR detail.
