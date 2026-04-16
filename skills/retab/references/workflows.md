# Workflows

Use this reference when the user already has a Retab workflow and needs to run it from code or poll for results.

## What this covers

- Start a workflow run with input documents
- Pass JSON into `start_json` blocks
- Poll a run until it finishes
- Read `final_outputs` from the completed run

This file does not cover visual workflow authoring in the dashboard.

## Input shape note

Workflow runs are the main REST exception in this skill:

- Direct document routes use `document: { filename, url }`
- Workflow run routes use `documents: { block_id: { filename, content, mime_type } }`

Do not reuse the direct-document payload shape when calling workflow runs over REST.

## Python

```python
from pathlib import Path
import time
from retab import Retab

client = Retab()

run = client.workflows.runs.create(
    workflow_id="wf_abc123",
    documents={
        "document-block-id": Path("invoice.pdf"),
    },
    json_inputs={
        "json-block-id": {"customer_id": "cust_123"},
    },
)

while run.status in ["pending", "running"]:
    time.sleep(1)
    run = client.workflows.runs.get(run.id)

print(run.status)
print(run.final_outputs)
```

## Node

```ts
import { Retab } from "@retab/node";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });

let run = await client.workflows.runs.create({
  workflowId: "wf_abc123",
  documents: {
    "document-block-id": "./invoice.pdf",
  },
  jsonInputs: {
    "json-block-id": { customerId: "cust_123" },
  },
});

while (run.status === "pending" || run.status === "running") {
  await new Promise((resolve) => setTimeout(resolve, 1000));
  run = await client.workflows.runs.get(run.id);
}

console.log(run.status);
console.log(run.finalOutputs);
```

## REST

Start a run:

```bash
curl -X POST "https://api.retab.com/v1/workflows/wf_abc123/run" \
  -H "Api-Key: $RETAB_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "documents": {
      "document-block-id": {
        "filename": "invoice.pdf",
        "content": "<BASE64_ENCODED_FILE_CONTENT>",
        "mime_type": "application/pdf"
      }
    },
    "json_inputs": {
      "json-block-id": { "customer_id": "cust_123" }
    }
  }'
```

Get a run:

```bash
curl -X GET "https://api.retab.com/v1/workflows/runs/run_abc123" \
  -H "Api-Key: $RETAB_API_KEY"
```

## Inputs

- `documents`: keys must match the workflow's document start block IDs
- `json_inputs`: keys must match the workflow's JSON start block IDs

## Outputs

- `status`: usually `pending`, `running`, `completed`, or `failed`
- `steps`: per-block execution details
- `final_outputs`: terminal outputs keyed by end block ID
- `error`: failure detail when the run does not succeed

## Guidance

- Use workflows when the same multi-step pipeline should run repeatedly.
- If the user only needs one operation, prefer the direct document routes instead.
- If the API reports missing input documents, the keys usually do not match the workflow's start block IDs.
- If the run fails after starting, inspect `steps`, `error`, and `final_outputs` before changing the input payload.
