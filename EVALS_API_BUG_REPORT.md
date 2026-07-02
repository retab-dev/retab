# Bug report: Workflow Evals API not functional (create returns 405, list returns 404)

**Date:** 2026-06-22
**Reporter:** anas.labgoul@retab.com
**Organization:** XPO (`org_01KHRJN4WXG316EENY8KF831VS`)
**Environment:** Production (`env_KbtPMIeRfiphTVy_KWhVJ`)
**Severity:** High — the Workflow Evals ("workflow tests") feature is completely unusable via API/CLI on this environment.

---

## Summary

The Workflow Evals endpoints under `/v1/workflows/evals` are advertised by the
CLI and the official SDKs, but the production API at `https://api.retab.com`
does not serve them for this organization/environment:

- `POST /v1/workflows/evals` (create an eval) → **405 Method Not Allowed**
- `GET  /v1/workflows/evals?workflow_id=...` (list evals) → **404 "Workflow not found"**

As a result it is impossible to create or even list workflow evals (the
"tests" section) programmatically. This is a **server-side** problem: the CLI
and SDKs send exactly the request the published API contract defines.

## Environment

- CLI: `retab 0.2.5 (commit f8ebebda9e2721ad7acb483c916c8bff24fc394a, built 2026-06-19T21:40:40Z)`
- OS: Windows 11
- API base URL: `https://api.retab.com`
- Auth: OAuth, organization scoped to XPO / Production (verified via `retab auth status`)

## Steps to reproduce

1. `retab org switch XPO`
2. Attempt to create an eval on the "XPO NEW - PRODUCTION" workflow:

   ```
   retab workflows evals create wrk_-dF0BUWtswiL_8w3XTj6o \
     --name "report-probe" \
     --block-id block_VDT-XDa6X0XqLb-Qm22PB \
     --run-id run_jrxEjv832am64um-_T9yV \
     --path orders --equals 1 --debug
   ```

3. Attempt to list evals:

   ```
   retab workflows evals list wrk_-dF0BUWtswiL_8w3XTj6o --debug
   ```

## Expected

- `create` returns `201`/`200` with the created eval object.
- `list` returns `200` with the (possibly empty) array of evals for the workflow.

## Actual (verbatim, from `--debug`)

**Create:**

```
POST /v1/workflows/evals HTTP/1.1
HTTP/2.0 405 Method Not Allowed
Allow: OPTIONS, GET, DELETE, PATCH
Date: Mon, 22 Jun 2026 10:31:45 GMT
X-Request-Id: req_1782124305244310718
Body: 405 method not allowed
```

The `Allow` header lists `OPTIONS, GET, DELETE, PATCH` — **`POST` is missing**
on the collection route `/v1/workflows/evals`.

**List:**

```
GET /v1/workflows/evals?limit=50&workflow_id=wrk_-dF0BUWtswiL_8w3XTj6o HTTP/1.1
HTTP/2.0 404 Not Found
Date: Mon, 22 Jun 2026 10:31:33 GMT
X-Request-Id: req_1782124333231079852
Body: {"detail":"Workflow not found"}
```

`404 "Workflow not found"` is returned even though the workflow exists and is
actively editable/publishable in the same CLI session (e.g.
`retab workflows blocks update`, `retab workflows publish`, `retab workflows runs list`
all succeed for `wrk_-dF0BUWtswiL_8w3XTj6o`).

## Request IDs (for server-side tracing)

- Create 405: `req_1782124305244310718` (2026-06-22T10:31:45Z)
- List 404:   `req_1782124333231079852` (2026-06-22T10:31:33Z)

## Evidence that the client side is correct (not a CLI bug)

Confirmed against the `retab` monorepo:

- The CLI command calls the SDK create method:
  `cli/cmd/zz_oagen_workflows_evals.go` → `client.Workflows.Evals.Create(...)`.
- The Go SDK manifest declares the operation:
  `clients/go/.oagen-manifest.json` → `"POST /v1/workflows/evals"`.
- The .NET SDK does the same:
  `clients/dotnet/src/Services/WorkflowEvals/WorkflowEvalsService.cs` →
  `PostAsync("/v1/workflows/evals")`.
- The feature is new on the client: the evals CLI files were first added on
  **2026-06-19** (commit `4759bc70`, "rename workflow tests to evals"), whereas
  the comparable (working) **experiments** feature was added 2026-05-14.

This points to a **client-ahead-of-server rollout**: the CLI/SDK shipped the
Workflow Evals client before the server enabled the endpoints on this
environment.

## Likely root cause (server-side)

On `api.retab.com` for this org/env, the `/v1/workflows/evals` router appears
to be only partially deployed:

- The collection `POST` (create) handler is not registered (hence `405` with an
  `Allow` set that omits `POST`).
- The collection `GET` (list) handler resolves but cannot find the workflow
  (returns `404 "Workflow not found"`), suggesting the evals service isn't
  correctly bound to the workflow store for this environment.

## Requested fix

1. **Primary (server):** Enable/deploy the Workflow Evals endpoints for this
   organization/environment — at minimum `POST /v1/workflows/evals` (create)
   and a working `GET /v1/workflows/evals?workflow_id=...` (list).
2. **Secondary (CLI UX):** When the server returns `405`/`404` for evals,
   surface a clearer message (e.g. "Workflow Evals are not enabled on this
   environment yet") instead of the raw `405 method not allowed` /
   `Workflow not found`, so users aren't misled into thinking the CLI is broken.

## Workaround in use

A local, deterministic regression harness (`xpo_workflow_tests/regression_tests.py`)
that pulls the live function-block code via the CLI and tests the merge/
single-commodity logic offline. It does not depend on the evals endpoints.
