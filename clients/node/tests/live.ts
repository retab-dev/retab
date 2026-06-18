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
import { probeLiveServer } from './live-preflight';

const API_KEY = process.env.RETAB_API_KEY;
const BASE_URL = process.env.RETAB_API_BASE_URL;

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
const PROBE = await probeLiveServer(API_KEY, BASE_URL);

/** True only when creds are present AND the configured server is reachable. */
export const LIVE = PROBE.live;

/** Reason string surfaced when an e2e suite is skipped. */
export const LIVE_SKIP_REASON = PROBE.reason;

/**
 * Non-empty when credentials/server are configured but unusable. The dedicated
 * live preflight e2e test fails on this, while other e2e files skip cleanly
 * instead of cascading import-initialization errors.
 */
export const LIVE_FATAL_REASON = !PROBE.live && PROBE.fatal ? PROBE.reason : '';

let cached: Retab | undefined;

/**
 * Returns a process-wide singleton live client. Throws if called when LIVE is
 * false — callers must gate on LIVE (e.g. `describe.skipIf(!LIVE)`) first.
 */
export function liveClient(): Retab {
  if (!LIVE) {
    throw new Error(
      LIVE_FATAL_REASON || LIVE_SKIP_REASON || 'liveClient() called without live e2e configuration'
    );
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
