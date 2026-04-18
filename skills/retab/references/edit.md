# Edit

Endpoint: `POST /v1/edits`

## Use it when

- You need to fill a form-like PDF or Office document
- You want to update a document from natural-language instructions
- You need the filled document returned as file data

## Minimal Python

```python
from retab import Retab

client = Retab()
result = client.edits.create(
    document="form.pdf",
    model="retab-small",
    instructions="Fill full name as John Doe and date of birth as 1990-01-15.",
)

print(result.id)
print(result.data.form_data)
print(result.data.filled_document)
```

## Minimal Node

```ts
import { Retab } from "@retab/node";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });

const result = await client.edits.create({
  document: "form.pdf",
  model: "retab-small",
  instructions: "Fill full name as John Doe and date of birth as 1990-01-15.",
});

console.log(result.id);
console.log(result.data.form_data);
console.log(result.data.filled_document);
```

## Minimal REST

```bash
curl -X POST "https://api.retab.com/v1/edits" \
  -H "Api-Key: $RETAB_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "filename": "form.pdf",
      "url": "data:application/pdf;base64,..."
    },
    "model": "retab-small",
    "instructions": "Fill full name as John Doe and date of birth as 1990-01-15."
  }'
```

## Useful fields

- `instructions`: required
- `form_fields`: optional for PDF flows where fields are already known
- `config`: edit rendering settings

## Response shape

- `id`: edit resource id
- `data.form_data[]`: field metadata and filled values
- `data.filled_document`: output file as MIME data

## Guidance

- Write explicit instructions with exact values.
- Use `form_fields` when field locations are already known and you want to skip inference.
- Save the returned `filled_document` if the user needs a file on disk.
