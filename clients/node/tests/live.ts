// Shared harness for REAL end-to-end tests that hit a live Retab server.
//
// Unlike the offline request-contract tests (which inject a fake `fetch` and
// never touch the network), the e2e tests built on this helper construct a real
// `Retab` client from RETAB_API_BASE_URL + RETAB_API_KEY and await live API
// calls. `setup.ts` (preloaded by bunfig.toml) has already loaded the matching
// .env file / process.env by the time this module is imported.
//
// CREDITLESS CONTRACT: helpers here only ever exercise list/get/upload/CRUD on
// config + storage resources. They never trigger LLM inference, document
// processing, or any billable operation, and they clean up only what they
// created.
//
// When RETAB_API_KEY / RETAB_API_BASE_URL are not set (e.g. a plain offline CI
// run), `LIVE` is false and e2e suites skip themselves via `describe.skipIf`
// so the offline contract tests keep passing standalone.

import { createHash } from 'crypto';

import { Retab } from '../src/index.js';

const API_KEY = process.env.RETAB_API_KEY;
const BASE_URL = process.env.RETAB_API_BASE_URL;

const HAVE_CREDS = Boolean(API_KEY && BASE_URL);

type ProbeResult = { live: true; reason: '' } | { live: false; reason: string; fatal: boolean };

/**
 * Probe the configured server once at module load. The e2e suites only run when
 * BOTH credentials are present AND the server actually answers — so a stale
 * `.env.local` pointing at a `localhost:4000` that isn't running (or any offline
 * CI run) cleanly skips the e2e suites instead of failing with connection
 * errors, keeping the offline contract tests green standalone. A reachable
 * server that rejects the configured API key is treated as a real e2e
 * misconfiguration and fails at module load with a concise message.
 *
 * Top-level await is supported by Bun's test module loader.
 */
async function probeServer(): Promise<ProbeResult> {
  if (!HAVE_CREDS) {
    return {
      live: false,
      fatal: false,
      reason: 'live e2e skipped: set RETAB_API_KEY and RETAB_API_BASE_URL to run against a server',
    };
  }
  try {
    const ctrl = new AbortController();
    const t = setTimeout(() => ctrl.abort(), 4000);
    const res = await fetch(`${BASE_URL}/v1/files?limit=1`, {
      method: 'GET',
      headers: { Authorization: `Bearer ${API_KEY}`, Accept: 'application/json' },
      signal: ctrl.signal,
    });
    clearTimeout(t);
    if (res.status >= 200 && res.status < 300) {
      return { live: true, reason: '' };
    }
    if (res.status === 401) {
      return {
        live: false,
        fatal: true,
        reason: 'live e2e preflight failed: RETAB_API_KEY was rejected by the configured server',
      };
    }
    return {
      live: false,
      fatal: true,
      reason: `live e2e preflight failed: ${BASE_URL}/v1/files?limit=1 returned HTTP ${res.status}: ${(await res.text()).slice(0, 240)}`,
    };
  } catch {
    return {
      live: false,
      fatal: false,
      reason:
        'live e2e skipped: configured server is unreachable (start the server or point RETAB_API_BASE_URL at one)',
    };
  }
}

const PROBE = await probeServer();
if (!PROBE.live && PROBE.fatal) {
  throw new Error(PROBE.reason);
}

/** True only when creds are present AND the configured server is reachable. */
export const LIVE = PROBE.live;

/** Reason string surfaced when an e2e suite is skipped. */
export const LIVE_SKIP_REASON = PROBE.reason;

let cached: Retab | undefined;

/**
 * Returns a process-wide singleton live client. Throws if called when LIVE is
 * false — callers must gate on LIVE (e.g. `describe.skipIf(!LIVE)`) first.
 */
export function liveClient(): Retab {
  if (!LIVE) {
    throw new Error('liveClient() called without RETAB_API_KEY/RETAB_API_BASE_URL');
  }
  if (!cached) {
    cached = new Retab({ apiKey: API_KEY as string, baseUrl: BASE_URL as string });
  }
  return cached;
}

/** A client built with a deliberately-invalid API key, for 401 assertions. */
export function bogusKeyClient(): Retab {
  return new Retab({ apiKey: 'sk_bogus_invalid_e2e_key', baseUrl: BASE_URL as string });
}

/** Unique-ish suffix so concurrently-created resources never collide. */
export function uniqueSuffix(): string {
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
}

/**
 * Upload a small in-memory PDF through the real 3-step file lifecycle
 * (create_upload -> signed PUT -> complete_upload) and return the file id.
 *
 * CREDITLESS: this is storage only. No parsing/OCR/extraction is triggered.
 * The signed PUT MUST send back exactly the `uploadHeaders` the server returns
 * — adding or overriding headers (e.g. an extra Content-Type) breaks the GCS
 * signature and yields a 403.
 */
export async function uploadTinyPdf(
  client: Retab,
  filename = `node-e2e-${uniqueSuffix()}.pdf`
): Promise<{ fileId: string; filename: string; sha256: string }> {
  const bytes = Buffer.from(`%PDF-1.4\n%node creditless e2e ${uniqueSuffix()}\n`, 'utf8');
  const sha256 = createHash('sha256').update(bytes).digest('hex');

  const created = await client.files.create_upload(
    filename,
    bytes.length,
    'application/pdf',
    sha256
  );

  const put = await fetch(created.uploadUrl, {
    method: created.uploadMethod,
    headers: created.uploadHeaders ?? {},
    body: bytes,
  });
  if (!put.ok) {
    throw new Error(`signed upload PUT failed: ${put.status} ${await put.text()}`);
  }

  await client.files.complete_upload(created.fileId, sha256);
  return { fileId: created.fileId, filename, sha256 };
}

/**
 * Discover a project id from existing workflows so creditless workflow CRUD can
 * attach to a real project. Returns null when the org has no workflows yet
 * (callers should skip rather than fabricate a project).
 */
export async function discoverProjectId(client: Retab): Promise<string | null> {
  const page = await client.workflows.list({ limit: 5 });
  for (const wf of page.data) {
    if (wf.projectId) return wf.projectId;
  }
  return null;
}
