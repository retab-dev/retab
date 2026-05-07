---
name: retab
description: Build apps and integrations on top of Retab's core APIs. Use when Codex needs to add document parsing, structured extraction, bundle splitting, form filling or document editing, document classification, or workflow run integration to a codebase through the Retab Python SDK, Node SDK, or direct REST calls. Covers starting workflow runs, waiting for completion, inspecting step executions, and handling human-review pauses.
---

# Retab

Use this skill to implement Retab document API calls and workflow runs without relying on external docs.

It covers:

- Direct document operations: `parse`, `extract`, `split`, `edit`, `classify`
- Existing workflow execution: start a run, pass inputs to start blocks, wait for completion, inspect step executions, and handle `waiting_for_human`

## Quick Start

1. Install a client when needed:
   - Python: `pip install retab`
   - Node: `npm install @retab/node`
2. Load `RETAB_API_KEY`.
3. Pick the smallest operation that solves the task:
   - Need text or page content: use `parse`
   - Need structured JSON from a schema: use `extract`
   - Need to break a file into labeled sections: use `split`
   - Need to fill or update a form-like document: use `edit`
   - Need to choose one label from known categories: use `classify`
   - Need to run an existing multi-step pipeline, wait for completion, inspect block outputs, or handle human review: use `workflows`

## Common Setup

- Python SDK: `pip install retab`
- Node SDK: `npm install @retab/node`
- REST base URL: `https://api.retab.com`
- Auth header for REST: `Api-Key: $RETAB_API_KEY`

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

## Working Rules

- Prefer the SDK unless the codebase is already built around raw HTTP.
- Pass `model="retab-small"` explicitly unless the user asks for a different tradeoff.
- Keep request bodies minimal. Add optional fields only when they solve a real problem.
- For direct document REST calls, send the `Api-Key` header and a `document` object with `filename` and `url`.
- For workflow-run REST calls, send `documents` keyed by start block ID, with `filename`, `content`, and `mime_type`.
- For workflow-run SDK calls, map inputs by start block ID exactly. Do not invent friendly aliases for block keys.
- For SDK calls, prefer passing a local file path when possible.
- Keep uploads focused. Trim or split overly large documents before sending them.
- Use generous timeouts for slow or multi-page documents.
- Add retries for transient network or 5xx failures. Do not blindly retry validation errors.
- If a workflow run must finish before downstream code proceeds, use SDK waiting helpers instead of hand-writing ad hoc polling when the SDK provides one.
- If a workflow stops at `waiting_for_human`, do not treat that as a generic failure. Surface it explicitly and inspect the relevant step or HIL decision state.
- When debugging workflow outputs, use `workflows.runs.steps.list(run_id)` as the batch primitive. Use `steps.get(run_id, block_id)` for a single step. Avoid looping `run.steps` with per-step `steps.get()` calls because that creates an N+1 anti-pattern.
- To retrieve the typed resource produced by an inference step, use `step.artifact` and the matching resource client: `client.extractions.get(step.artifact.id)`, `client.splits.get(...)`, `client.classifications.get(...)`, `client.parses.get(...)`, `client.edits.get(...)`, or `client.partitions.get(...)`.
- To retrieve source provenance for an extraction, use `client.extractions.sources(extraction.id)` or `GET /v1/extractions/{extraction_id}/sources`; the returned `sources` tree mirrors the extraction and wraps leaves as `{ value, source }`.
- Stay within this skill's scope. It covers direct document routes plus running existing workflows. If the user asks for workflow design, widgets, projects, or MCP setup, give the simplest useful answer and note that those areas are outside this skill's main coverage.

## Operation Chooser

- `parse`: convert a document into page-by-page text or structured table output
- `extract`: map a document into a JSON schema
- `split`: assign pages to named subdocuments
- `edit`: fill or update a form-like document from natural-language instructions
- `classify`: choose one category from a fixed list
- `workflows`: run an existing multi-step workflow and poll its outputs

Choose `workflows` instead of the direct routes when:

- The user already has a workflow ID such as `wf_...`
- The user already has a dashboard workflow
- The pipeline involves multiple operations or conditional branching
- You need end-block `final_outputs` rather than one direct API response
- You need step-by-step inspection, retries, or human-review checkpoints
- They mention `final_outputs`, `steps`, `waiting_for_human`, or HIL/human review

Use a direct route when the user only needs one operation and no saved workflow exists.

## REST Skeleton

Direct document routes use this shape:

```bash
curl -X POST "https://api.retab.com/v1/RESOURCE_ROUTE" \
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

Use the actual route:

- Parse: `/v1/parses`
- Extract: `/v1/extractions`
- Split: `/v1/splits`
- Edit: `/v1/edits`
- Classify: `/v1/classifications`

Workflow runs are the main exception to this skeleton. They use `documents` keyed by start block ID. See the workflow section below.

## Parse

Endpoint: `POST /v1/parses`

Use it when:

- You need readable text for prompting, indexing, or debugging
- You want page-by-page output instead of schema extraction
- You need table output in `html`, `markdown`, `yaml`, or `json`
- You want free text, not structured field extraction
- You want the result persisted and retrievable later via `client.parses.get(id)` or `client.parses.list()`

Minimal Python:

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

Minimal Node:

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

Minimal REST:

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

Useful fields:

- `table_parsing_format`: defaults to `html`
- `image_resolution_dpi`: defaults to `192`
- `instructions`: optional free-form instructions for domain hints or layout guidance
- Parse is non-streaming in the SDK
- The documented SDK surface is `document`, `model`, `table_parsing_format`, `image_resolution_dpi`, `instructions`

Response shape:

- `id`: parse resource id
- `file.id`, `file.filename`, `file.mime_type`
- `output.pages`: page-by-page parsed content
- `output.text`: full document content concatenated
- `model`, `table_parsing_format`, `image_resolution_dpi`
- `usage.page_count`, `usage.credits`
- `created_at`, `updated_at`

Listing and retrieving parses:

```python
parse = client.parses.get("parse_01G34H8J2K")
for item in client.parses.list(limit=20).data:
    ...
client.parses.delete(parse.id)
```

Guidance:

- Use `parses.create` when the task only needs free text.
- Use `extract` directly when you need schema-shaped output.
- Use `table_parsing_format` when downstream code needs predictable table output.
- Start at `192` DPI, lower to `96` for speed, and raise toward `300` for hard OCR cases.
- Do not rely on unsupported parse fields like `browser_canvas`.

## Extract

Endpoint: `POST /v1/extractions`

Use it when:

- You need structured JSON back from a document
- The target shape is known or can be written as JSON Schema
- You want a persisted extraction resource you can fetch, list, delete, or source later

If the user wants structured extraction but does not yet have a schema:

- Write a small schema manually when the output is obvious
- Otherwise, generate or draft a schema before implementing `extract`
- Do not pretend `extract` can infer arbitrary output structure without a schema

Minimal Python:

```python
from retab import Retab

client = Retab()
response = client.extractions.create(
    document="invoice.pdf",
    model="retab-small",
    json_schema={
        "type": "object",
        "properties": {
            "invoice_number": {"type": "string"},
            "total_amount": {"type": "number"},
        },
        "required": ["invoice_number", "total_amount"],
    },
)

print(response.id)
print(response.output)
```

Minimal Node:

```ts
import { Retab } from "@retab/node";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });

const response = await client.extractions.create({
  document: "invoice.pdf",
  model: "retab-small",
  json_schema: {
    type: "object",
    properties: {
      invoice_number: { type: "string" },
      total_amount: { type: "number" },
    },
    required: ["invoice_number", "total_amount"],
  },
});

console.log(response.id);
console.log(response.output);
```

Minimal REST:

```bash
curl -X POST "https://api.retab.com/v1/extractions" \
  -H "Api-Key: $RETAB_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "filename": "invoice.pdf",
      "url": "data:application/pdf;base64,..."
    },
    "model": "retab-small",
    "json_schema": {
      "type": "object",
      "properties": {
        "invoice_number": { "type": "string" },
        "total_amount": { "type": "number" }
      },
      "required": ["invoice_number", "total_amount"]
    }
  }'
```

Useful fields:

- `json_schema`: required
- `n_consensus`: defaults to `1`; raise it when accuracy matters more than speed
- `image_resolution_dpi`: defaults to `192`
- `instructions`: optional free-form instructions to steer the extraction
- `additional_messages`: append raw chat messages to the LLM request when you need full message-level control

Consensus:

- `n_consensus=1`: one extraction pass, cheapest and fastest
- `n_consensus>1`: multiple extraction passes, then field-by-field reconciliation
- Consensus improves robustness when fields are ambiguous, noisy, or hard to read
- Consensus produces more useful `likelihoods`, because agreement across runs becomes a signal of confidence

Retab reconciles values by type:

- Scalars such as booleans and exact strings lean on voting or matching
- Numbers are grouped by closeness rather than exact equality
- Nested objects are reconciled field by field
- Arrays are reconciled element by element after alignment

Raise `n_consensus` when:

- The extraction is business-critical
- The document quality is poor
- The schema contains ambiguous fields
- You plan to gate downstream automation on confidence

Keep `n_consensus=1` when:

- You need low latency
- The fields are simple and stable
- You are still iterating on the schema and want quick feedback

`consensus.likelihoods` mirrors the extracted structure and reflects how strongly the runs agreed on each field. If a field keeps getting low likelihoods, first improve the schema before just increasing `n_consensus`. Better field names, tighter descriptions, and stricter types usually help more than extra retries.

Response shape:

- `id`: extraction resource id
- `output`: extracted JSON
- `consensus.likelihoods`: per-field confidence values
- `usage`: token and cost-related metadata

Sources:

```python
sources = client.extractions.sources(response.id)

print(sources.extraction)
print(sources.sources["invoice_number"]["source"])
```

```ts
const sources = await client.extractions.sources(response.id);

console.log(sources.extraction);
console.log(sources.sources.invoice_number.source);
```

Raw REST equivalent:

```bash
curl -X GET "https://api.retab.com/v1/extractions/${EXTRACTION_ID}/sources" \
  -H "Api-Key: $RETAB_API_KEY"
```

The sources response includes:

- `extraction`: the original extracted output
- `sources`: a tree mirroring the extraction where each leaf is wrapped as `{ value, source }`
- `source`: citation content, surrounding context, and a format-specific anchor when available

Guidance:

- If no schema exists yet, draft one before implementing `extract`.
- Keep schemas small and explicit. Overly broad schemas reduce reliability.
- Use `extract` only when the output must be structured. Otherwise use `parse`.

## Split

Endpoint: `POST /v1/splits`

Use it when:

- One uploaded file contains multiple subdocuments
- You need page assignments for each subdocument type
- The same subdocument can appear multiple times in one file
- You want to route each detected subdocument into a different downstream step

Do not use `split` for key-based grouping inside one homogeneous document type. Use `POST /v1/partitions` for that.

Minimal Python:

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

Minimal Node:

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

Minimal REST:

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

Useful fields:

- `subdocuments`: required; each item needs `name`, optional `description`, and optional `allow_multiple_instances`
- `instructions`: optional free-form instructions to steer the split
- `n_consensus`: raise it when split boundaries are business-critical
- `bust_cache`: force a fresh run

Response shape:

- `output[].name`
- `output[].pages`
- `consensus.likelihoods[]` with `name` and `pages` when `n_consensus > 1`
- `consensus.choices[]` as raw voter outputs when `n_consensus > 1`

Guidance:

- Write distinct subdocument descriptions. Overlapping labels make routing worse.
- Use `allow_multiple_instances` when one subdocument type can repeat.
- Use `split` before `extract` when a bundle must be separated first.
- Use `partition` after `split` if a specific split output then needs key-based grouping.

## Edit

Endpoint: `POST /v1/edits`

Use it when:

- You need to fill a form-like PDF or Office document
- You want to update a document from natural-language instructions
- You need the filled document returned as file data

Minimal Python:

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

Minimal Node:

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

Minimal REST:

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

Useful fields:

- `instructions`: required
- `form_fields`: optional for PDF flows where fields are already known
- `config`: edit rendering settings

Response shape:

- `id`: edit resource id
- `data.form_data[]`: field metadata and filled values
- `data.filled_document`: output file as MIME data

Guidance:

- Write explicit instructions with exact values.
- Use `form_fields` when field locations are already known and you want to skip inference.
- Save the returned `filled_document` if the user needs a file on disk.

## Classify

Endpoint: `POST /v1/classifications`

Use it when:

- You need one label from a fixed set of categories
- You want a lightweight routing step before more expensive processing
- The result can be decided from document semantics rather than exact field extraction

Minimal Python:

```python
from retab import Retab

client = Retab()
result = client.classifications.create(
    document="document.pdf",
    model="retab-small",
    n_consensus=3,
    categories=[
        {"name": "invoice", "description": "Invoice documents"},
        {"name": "receipt", "description": "Receipt documents"},
        {"name": "contract", "description": "Contract documents"},
    ],
)

print(result.output.category)
print(result.output.reasoning)
```

Minimal Node:

```ts
import { Retab } from "@retab/node";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });

const result = await client.classifications.create({
  document: "document.pdf",
  model: "retab-small",
  categories: [
    { name: "invoice", description: "Invoice documents" },
    { name: "receipt", description: "Receipt documents" },
    { name: "contract", description: "Contract documents" },
  ],
});

console.log(result.output.category);
console.log(result.output.reasoning);
console.log(result.consensus?.likelihood);
```

Minimal REST:

```bash
curl -X POST "https://api.retab.com/v1/classifications" \
  -H "Api-Key: $RETAB_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "document": {
      "filename": "document.pdf",
      "url": "data:application/pdf;base64,..."
    },
    "model": "retab-small",
    "n_consensus": 3,
    "categories": [
      { "name": "invoice", "description": "Invoice documents" },
      { "name": "receipt", "description": "Receipt documents" },
      { "name": "contract", "description": "Contract documents" }
    ]
  }'
```

Useful fields:

- `categories`: required; each item needs `name` and `description`
- `first_n_pages`: limit analysis when early pages are enough
- `instructions`: optional free-form instructions for ambiguous categories, such as "invoices with totals under $100 should be classified as receipts"
- `n_consensus`: run classification multiple times and majority-vote the final label

Response shape:

- `output.category`
- `output.reasoning`
- `consensus.likelihood` when consensus produced at least two successful votes
- `consensus.choices[]` with the individual classification votes when `n_consensus > 1`

Guidance:

- Keep categories mutually exclusive when possible.
- Use `first_n_pages` when the label can be determined early and latency matters.
- Use `classify` for routing. Use `extract` when you need structured values.

## Workflows

Use workflows when the user already has a Retab workflow and needs to run it from code, wait for results, inspect step executions, or handle a paused human-review state.

This covers:

- Start a workflow run with input documents
- Pass JSON into `start_json` blocks
- Wait for a run until it reaches a terminal state
- Read `final_outputs` from the completed run
- Inspect `steps` and fetch per-block execution records
- Handle `waiting_for_human` explicitly
- Choose between direct-route APIs and workflow execution

This does not cover visual workflow authoring in the dashboard.

Use a workflow when:

- The user already has pipeline logic in Retab and wants to reuse it from code
- You need to run one saved workflow on many documents
- You need to chain parse, classify, split, extract, edit, or custom logic without rebuilding each step in app code
- You need intermediate block outputs, not just the final result
- The workflow can pause for human review as part of the run lifecycle

If the user only needs one direct operation and no saved workflow exists, prefer the direct resource routes.

### Workflow Input Shape

Workflow runs are the main REST exception:

- Direct document routes use `document: { filename, url }`
- Workflow run routes use `documents: { block_id: { filename, content, mime_type } }`

Do not reuse the direct-document payload shape when calling workflow runs over REST.

Block ID rules:

- `documents` keys must match document start block IDs exactly
- `json_inputs` keys must match `start_json` block IDs exactly
- If the API says an input is missing, the key name is often wrong even when the file itself is valid
- When a workflow has multiple start blocks, pass only the blocks you actually need, but use the real IDs

### Workflow Status Model

Expect these workflow run statuses most often:

- `pending` or `queued`: accepted but not started yet
- `running`: actively executing
- `completed`: finished successfully
- `error`: failed during execution
- `cancelled`: stopped before completion
- `waiting_for_human`: paused and waiting for a human decision

Treat `waiting_for_human` as its own outcome. It is not a generic failure and often requires fetching step or HIL decision details instead of retrying automatically.

### Workflow Python

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

run.raise_for_status()
print(run.status)
print(run.final_outputs)
```

Python with SDK waiting helper:

```python
from pathlib import Path
from retab import Retab

client = Retab()

run = client.workflows.runs.create(
    workflow_id="wf_abc123",
    documents={"document-block-id": Path("invoice.pdf")},
)
run = client.workflows.runs.wait_for_completion(
    run.id,
    poll_interval_seconds=2.0,
    timeout_seconds=600.0,
)

if run.status == "waiting_for_human":
    print("Run paused for human review")
else:
    run.raise_for_status()
    print(run.final_outputs)
```

### Workflow Node

```ts
import { Retab } from "@retab/node";
import { raiseForStatus } from "retab";

const client = new Retab({ apiKey: process.env.RETAB_API_KEY });

const run = await client.workflows.runs.createAndWait({
  workflowId: "wf_abc123",
  documents: {
    "document-block-id": "./invoice.pdf",
  },
  jsonInputs: {
    "json-block-id": { customerId: "cust_123" },
  },
  pollIntervalMs: 2000,
  timeoutMs: 600000,
});

if (run.status === "waiting_for_human") {
  console.log("Run paused for human review");
} else {
  raiseForStatus(run);
  console.log(run.finalOutputs);
}
```

Node manual polling:

```ts
let run = await client.workflows.runs.create({
  workflowId: "wf_abc123",
  documents: { "document-block-id": "./invoice.pdf" },
});

run = await client.workflows.runs.waitForCompletion(run.id, {
  pollIntervalMs: 2000,
  timeoutMs: 600000,
  onStatus: (r) => console.log(`status=${r.status}`),
});
```

### Workflow REST

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

Inputs:

- `documents`: keys must match the workflow's document start block IDs
- `json_inputs`: keys must match the workflow's JSON start block IDs

Outputs:

- `status`: usually `pending`, `running`, `completed`, or `failed`
- `steps`: per-block execution details
- `final_outputs`: terminal outputs keyed by end block ID
- `error`: failure detail when the run does not succeed

### Inspecting Step Executions

Top-level `final_outputs` is useful for end results, but workflows often need step-level inspection for debugging or partial consumption. For every step in one call, use `steps.list(run_id)`. It returns the full persisted step list in a single request.

Python:

```python
# Batch, one HTTP call for the whole run:
for step in client.workflows.runs.steps.list(run.id):
    print(step.block_id, step.status, step.error, step.artifact)

# Single step:
step = client.workflows.runs.steps.get(run.id, "extract-block-1")
print(step.status, step.error)
print(step.extracted_data)  # handle-derived shortcut

# Jump to the typed underlying resource:
if step.artifact:
    extraction = client.extractions.get(step.artifact.id)
    # equivalents: client.splits.get / classifications.get / parses.get / edits.get / partitions.get
```

Node:

```ts
for (const step of await client.workflows.runs.steps.list(run.id)) {
  console.log(step.block_id, step.status, step.error, step.artifact);
}

const step = await client.workflows.runs.steps.get(run.id, "extract-block-1");
if (step.artifact) {
  const extraction = await client.extractions.get(step.artifact.id);
  console.log(extraction);
}
```

Every executed block exposes a primary `step.artifact` `{operation, id}` pointer. Inference blocks point at their typed resource; other block types point at a workflow-native block artifact.

| `step.artifact.operation` | Emitted by block type | Fetch with |
|---|---|---|
| `extraction` | `extract` | `client.extractions.get(id)` |
| `split` | `split` | `client.splits.get(id)` |
| `classification` | `classifier` | `client.classifications.get(id)` |
| `parse` | `parse` | `client.parses.get(id)` |
| `edit` | `edit` | `client.edits.get(id)` |
| `partition` | `for_each_sentinel_start` | `client.partitions.get(id)` |

`step.artifacts` contains every artifact ref for the block, primary first. For inference blocks, the first artifact is the typed domain resource.

Use step inspection when:

- `final_outputs` is empty or too coarse
- One intermediate block is failing or producing bad data
- You need output from a non-terminal block
- The workflow paused at `waiting_for_human` and you need the relevant block context

Workflow debugging guidance:

- If run creation fails immediately, check block ID keys first.
- If the run reaches `error`, inspect `run.error`, `run.steps`, and step executions before changing the input payload.
- If the run reaches `waiting_for_human`, fetch the relevant step execution or HIL decision state instead of retrying blindly.
- If only one block matters, fetch that block directly with `steps.get(...)`.
- If you need a snapshot of the entire execution, use `steps.list(...)` / `list(...)`.
- Prefer SDK waiting helpers over handwritten polling loops when they exist.
