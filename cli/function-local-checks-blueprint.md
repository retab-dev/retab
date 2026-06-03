# Blueprint: Best-Effort Local Function Checks

## Goal

`retab workflows blocks validate-config` should run local language checks for function bundles as a non-blocking bonus before or alongside the existing remote config validation.

The command remains primarily a config validation command. Missing local developer tools or local lint/typecheck failures should inform the user, not prevent remote validation or force a non-zero exit.

## Contract

`validate-config` should only fail for:

- Invalid or unreadable bundle files.
- Manifest/target mismatches.
- Unsupported block config shape.
- Remote validation errors when remote validation is enabled.
- CLI/runtime errors unrelated to optional local checks.

`validate-config` should not fail solely because:

- `tsc`, `ruff`, or `pyright` is not installed.
- A local typecheck/lint command exits non-zero.
- A local checker cannot run because the current bundle lacks enough generated local support.

Local checker outcomes are diagnostics. They should be visible in output and stderr/table rendering, but they are not authoritative.

## Output Shape

Add a `local_checks` array to `validate-config` output.

```json
{
  "ok": true,
  "mode": "remote",
  "authoritative": true,
  "workflow_id": "wrk_...",
  "block_id": "block_...",
  "block_type": "function",
  "adapter": "function",
  "config_hash": "server-hash",
  "local_checks": [
    {
      "name": "tsc",
      "language": "typescript",
      "status": "passed",
      "required": false
    },
    {
      "name": "pyright",
      "language": "python",
      "status": "skipped",
      "required": false,
      "message": "pyright is not installed"
    }
  ]
}
```

Statuses:

- `passed`: tool was found and exited zero.
- `failed`: tool was found and exited non-zero.
- `skipped`: tool was not found, the bundle language does not use it, or required support files are missing.

Each check may include:

- `command`: command line executed, with bundle-relative paths where possible.
- `exit_code`: process exit code when available.
- `stdout`: short captured output.
- `stderr`: short captured output.
- `message`: concise human-readable reason.
- `install_hint`: concise installation guidance for missing tools.

Output must be bounded. Truncate each stdout/stderr field to a small fixed limit, for example 8 KiB.

## Default Behavior

Default mode is best-effort auto:

```bash
retab workflows blocks validate-config <workflow-id> <block-id> --dir tmp/fn
```

Flow:

1. Read and reassemble the bundle.
2. Validate manifest and local config structure.
3. Run applicable local checks best-effort.
4. Run remote config validation unless `--offline` is set.
5. Return zero when bundle validation and remote validation pass, regardless of local check failures/skips.

## Local Check Selection

For TypeScript function bundles:

- Prefer `tsc --noEmit --pretty false`.
- Run from the bundle directory.
- Use the generated `tsconfig.json`.
- If `tsconfig.json` is missing, skip with a repair hint:
  `retab workflows blocks functions hydrate <dir> --force`.

For Python function bundles:

- Run `ruff check function.py` when `ruff` is available.
- Run `pyright function.py` when `pyright` is available.
- Keep both optional.
- If generated local stubs are too loose for useful Pyright output, still report the run result honestly; improving stubs is a separate task.

## Tool Discovery

Use PATH by default:

- `tsc`
- `ruff`
- `pyright`

Allow environment overrides:

- `RETAB_TSC`
- `RETAB_RUFF`
- `RETAB_PYRIGHT`

If an override is set and the executable cannot run, report `skipped` with the override path in the message. Do not fail the command.

## Hydration Additions

TypeScript hydration should write:

- `function.ts`
- `models.generated.ts`
- `schemas.generated.ts`
- `tsconfig.json`
- `run.mjs`
- `.retab/runtime.mjs`

Python hydration should continue to write:

- `function.py`
- `input.py`
- `output.py`
- `run.py`
- `.retab/runtime.py`

`doctor-config` should warn when optional local check tools are missing, but it should not mark the bundle unhealthy solely for missing optional tools.

## UX

For human/table output, show a compact section:

```text
Local checks:
  tsc      passed
  ruff     skipped  ruff is not installed
  pyright  failed   2 diagnostics
```

For JSON output, include full bounded diagnostic fields.

Do not bury remote validation status under local checks. Remote validation remains the authoritative result.

## Future Strict Mode

Do not add strict behavior by default.

A later CI-oriented flag can make local checks blocking:

```bash
--local-checks=auto|required|off
```

Initial implementation can hardcode `auto` behavior and leave this flag for a follow-up if needed.

If/when strict mode exists:

- `auto`: default, non-blocking.
- `off`: skip local checks.
- `required`: fail when an applicable tool is missing or exits non-zero.

## Test Plan

Unit tests:

- TypeScript function bundle with `tsc` missing reports `skipped` and exits zero when remote validation passes.
- TypeScript function bundle with fake passing `tsc` reports `passed`.
- TypeScript function bundle with fake failing `tsc` reports `failed` and still exits zero when remote validation passes.
- Python function bundle reports independent `ruff` and `pyright` statuses.
- Environment overrides are honored.
- `--offline` still runs local checks and reports `authoritative: false`.
- Non-function bundles return no local checks or an empty array.

Integration smoke:

- Pull a live TypeScript function block.
- Edit `function.ts`.
- Run `validate-config`; confirm local checks are reported and remote validation still runs.
- Run `push-config`; confirm unchanged drift-safety behavior.
