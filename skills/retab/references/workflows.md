# Workflows

Use this reference when the user already has a Retab workflow and needs to run it from code, wait for results, inspect step executions, or handle a paused human-review state.

## What this covers

- Start a workflow run with input documents
- Pass JSON into `start_json` blocks
- Wait for a run until it reaches a terminal state
- Read `final_outputs` from the completed run
- Inspect `steps` and fetch per-block execution records
- Handle `waiting_for_human` explicitly
- Choose between direct-route APIs and workflow execution

This file does not cover visual workflow authoring in the dashboard.

## When to use workflows

Use a workflow when the user already has pipeline logic in Retab and wants to reuse it from code.

Common cases:

- Run one saved workflow on many documents
- Chain parse, classify, split, extract, edit, or custom logic without rebuilding each step in app code
- Reuse dashboard-authored logic from an SDK or backend service
- Inspect intermediate block outputs, not just the final result
- Pause for human review as part of the run lifecycle

If the user only needs one direct operation and no saved workflow exists, prefer the direct resource routes instead of wrapping that work in a workflow call.

## Input shape note

Workflow runs are the main REST exception in this skill:

- Direct document routes use `document: { filename, url }`
- Workflow run routes use `documents: { block_id: { filename, content, mime_type } }`

Do not reuse the direct-document payload shape when calling workflow runs over REST.

## Block ID rules

- `documents` keys must match document start block IDs exactly
- `json_inputs` keys must match `start_json` block IDs exactly
- If the API says an input is missing, the key name is often wrong even when the file itself is valid
- When a workflow has multiple start blocks, pass only the blocks you actually need, but use the real IDs

## Status model

Expect these workflow run statuses most often:

- `pending` or `queued`: accepted but not started yet
- `running`: actively executing
- `completed`: finished successfully
- `error`: failed during execution
- `cancelled`: stopped before completion
- `waiting_for_human`: paused and waiting for a human decision

Treat `waiting_for_human` as its own outcome. It is not a generic failure and often requires fetching step or HIL decision details instead of retrying automatically.

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

run.raise_for_status()
print(run.status)
print(run.final_outputs)
```

### Python with SDK waiting helper

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

## Node

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

### Node manual polling

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

## Inspecting Step Executions

Top-level `final_outputs` is useful for end results, but workflows often need step-level inspection for debugging or partial consumption. For "give me every step in one call" use `steps.list(run_id)` — it returns the full persisted step list in a single request.

### Python step inspection

```python
# Batch (one HTTP call for the whole run):
for step in client.workflows.runs.steps.list(run.id):
    print(step.block_id, step.status, step.error, step.artifact)

# Single step:
step = client.workflows.runs.steps.get(run.id, "extract-block-1")
print(step.status, step.error)
print(step.extracted_data)   # handle-derived shortcut

# Jump to the typed underlying resource:
if step.artifact:
    extraction = client.extractions.get(step.artifact.id)
    # equivalents: client.splits.get / classifications.get / parses.get / edits.get / partitions.get
```

### Node step inspection

```ts
// Batch:
for (const step of await client.workflows.runs.steps.list(run.id)) {
  console.log(step.block_id, step.status, step.error, step.artifact);
}

// Single step + typed resource fetch:
const step = await client.workflows.runs.steps.get(run.id, "extract-block-1");
if (step.artifact) {
  const extraction = await client.extractions.get(step.artifact.id);
  console.log(extraction);
}
```

### `step.artifact`

Every executed block exposes a primary `step.artifact` `{operation, id}` pointer. Inference blocks point at their typed resource; other block types point at a workflow-native block artifact.

| `step.artifact.operation` | emitted by block type | fetch with |
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

## Debugging guidance

- If run creation fails immediately, check block ID keys first
- If the run reaches `error`, inspect `run.error`, `run.steps`, and step executions before changing the input payload
- If the run reaches `waiting_for_human`, fetch the relevant step execution or HIL decision state instead of retrying blindly
- If only one block matters, fetch that block directly with `steps.get(...)`
- If you need a snapshot of the entire execution, use `steps.get_all(...)` / `getAll(...)`

## Guidance

- Use workflows when the same multi-step pipeline should run repeatedly.
- If the user only needs one operation, prefer the direct document routes instead.
- If the API reports missing input documents, the keys usually do not match the workflow's start block IDs.
- If the run fails after starting, inspect `steps`, `error`, and `final_outputs` before changing the input payload.
- Prefer the SDK waiting helpers over handwritten polling loops when they exist.
