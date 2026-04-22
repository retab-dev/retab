---
name: retab
description: Build apps and integrations on top of Retab's core APIs. Use when Codex needs to add document parsing, structured extraction, bundle splitting, form filling or document editing, document classification, or workflow run integration to a codebase through the Retab Python SDK, Node SDK, or direct REST calls. Covers starting workflow runs, waiting for completion, inspecting step outputs, and handling human-review pauses.
---

# Retab

Use this skill to implement Retab document API calls and workflow runs without relying on external docs.

It is especially useful for two kinds of tasks:

- Direct document operations: `parse`, `extract`, `split`, `edit`, `classify`
- Existing workflow execution: start a run, pass inputs to start blocks, wait for completion, inspect step outputs, and handle `waiting_for_human`

## Quick Start

1. Install a client when needed:
   - Python: `pip install retab`
   - Node: `npm install @retab/node`
2. Load `RETAB_API_KEY`.
3. Pick the smallest operation that solves the task:
  - Need text or page content: read `references/parse.md`
  - Need structured JSON from a schema: read `references/extract.md`
  - Need to break a file into labeled sections: read `references/split.md`
  - Need to fill or update a form-like document: read `references/edit.md`
  - Need to choose one label from known categories: read `references/classify.md`
  - Need to run an existing multi-step pipeline, wait for completion, inspect block outputs, or handle human review: read `references/workflows.md`
4. Read `references/common.md` before writing code if authentication, input format, or model defaults are still unclear.

## Workflow-first cases

Prefer the workflow reference over direct document routes when the user already has a workflow and any of these are true:

- They mention a workflow ID such as `wf_...`
- They need multiple steps chained together
- They need outputs from specific workflow blocks
- They mention `final_outputs`, `steps`, `waiting_for_human`, or HIL/human review
- They want to reuse an existing dashboard workflow from code instead of rebuilding the logic inline

## Working Rules

- Prefer the SDK unless the codebase is already built around raw HTTP.
- Pass `model="retab-small"` explicitly unless the user asks for a different tradeoff.
- Keep request bodies minimal. Add optional fields only when they solve a real problem.
- For direct document REST calls, send the `Api-Key` header and a `document` object with `filename` and `url`.
- For workflow-run REST calls, send `documents` keyed by start block ID, with `filename`, `content`, and `mime_type`.
- For workflow-run SDK calls, map inputs by start block ID exactly. Do not invent friendly aliases for block keys.
- For SDK calls, prefer passing a local file path when possible.
- Add retries for transient network or 5xx failures. Do not blindly retry validation errors.
- If a workflow run must finish before downstream code proceeds, use the SDK waiting helpers instead of hand-writing ad hoc polling when the SDK already provides one.
- If a workflow stops at `waiting_for_human`, do not treat that as a generic failure. Surface it explicitly and inspect the relevant step or HIL decision state.
- When debugging workflow outputs, use `workflows.runs.steps.list(run_id)` as the batch primitive (one HTTP call for the whole run). Use `steps.get(run_id, block_id)` for a single step. Avoid looping `run.steps` with per-step `steps.get()` calls — that creates an N+1 anti-pattern.
- To retrieve the typed resource produced by an inference step (extract, split, classifier, parse, edit, for-each partition), use `step.artifact` and the matching resource client (`client.extractions.get(step.artifact.id)`, `client.splits.get(...)`, `client.classifications.get(...)`, `client.parses.get(...)`, `client.edits.get(...)`, `client.partitions.get(...)`).
- Stay within this skill's scope. It covers the direct document routes plus running existing workflows. If the user asks for workflow design, widgets, projects, or MCP setup, give the simplest useful answer and note that those areas are outside this skill's main coverage.

## References

- `references/common.md`: auth, SDK setup, shared request conventions, operation chooser
- `references/parse.md`: `POST /v1/parses`
- `references/extract.md`: `POST /v1/extractions`
- `references/split.md`: `POST /v1/splits`
- `references/edit.md`: `POST /v1/edits`
- `references/classify.md`: `POST /v1/classifications`
- `references/workflows.md`: workflow runs with `client.workflows.runs.create()` and `client.workflows.runs.get()`
