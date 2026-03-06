---
name: retab
description: Build apps and integrations on top of Retab's core APIs. Use when Codex needs to add document parsing, structured extraction, bundle splitting, form filling or document editing, document classification, or workflow run integration to a codebase through the Retab Python SDK, Node SDK, or direct REST calls.
---

# Retab

Use this skill to implement Retab document API calls and workflow runs without relying on external docs.

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
   - Need to run an existing workflow and poll for results: read `references/workflows.md`
4. Read `references/common.md` before writing code if authentication, input format, or model defaults are still unclear.

## Working Rules

- Prefer the SDK unless the codebase is already built around raw HTTP.
- Pass `model="retab-small"` explicitly unless the user asks for a different tradeoff.
- Keep request bodies minimal. Add optional fields only when they solve a real problem.
- For REST calls, send the `Api-Key` header and a `document` object with `filename` and `url`.
- For SDK calls, prefer passing a local file path when possible.
- Add retries for transient network or 5xx failures. Do not blindly retry validation errors.
- Stay within this skill's scope. It covers the direct document routes plus running existing workflows. If the user asks for workflow design, widgets, projects, or MCP setup, give the simplest useful answer and note that those areas are outside this skill's main coverage.

## References

- `references/common.md`: auth, SDK setup, shared request conventions, operation chooser
- `references/parse.md`: `POST /v1/documents/parse`
- `references/extract.md`: `POST /v1/documents/extract`
- `references/split.md`: `POST /v1/documents/split`
- `references/edit.md`: `POST /v1/documents/edit`
- `references/classify.md`: `POST /v1/documents/classify`
- `references/workflows.md`: workflow runs with `client.workflows.runs.create()` and `client.workflows.runs.get()`
