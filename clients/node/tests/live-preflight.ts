export type ProbeResult =
  | { live: true; reason: '' }
  | { live: false; reason: string; fatal: boolean };

type FetchLike = (input: string, init: RequestInit) => Promise<Response>;

const MISSING_CREDS_REASON =
  'live e2e skipped: set RETAB_API_KEY and RETAB_API_BASE_URL to run against a server';
const UNREACHABLE_REASON =
  'live e2e skipped: configured server is unreachable (start the server or point RETAB_API_BASE_URL at one)';
const INVALID_KEY_REASON =
  'live e2e preflight failed: RETAB_API_KEY was rejected by the configured server';

function normalizedBaseUrl(baseUrl: string): string {
  return baseUrl.replace(/\/+$/, '');
}

export async function probeLiveServer(
  apiKey: string | undefined,
  baseUrl: string | undefined,
  fetchImpl: FetchLike = fetch
): Promise<ProbeResult> {
  if (!apiKey || !baseUrl) {
    return { live: false, fatal: false, reason: MISSING_CREDS_REASON };
  }

  const url = `${normalizedBaseUrl(baseUrl)}/v1/files?limit=1`;
  const ctrl = new AbortController();
  const timeout = setTimeout(() => ctrl.abort(), 4000);
  try {
    const res = await fetchImpl(url, {
      method: 'GET',
      headers: { Authorization: `Bearer ${apiKey}`, Accept: 'application/json' },
      signal: ctrl.signal,
    });

    if (res.status >= 200 && res.status < 300) {
      return { live: true, reason: '' };
    }
    if (res.status === 401) {
      return { live: false, fatal: true, reason: INVALID_KEY_REASON };
    }
    return {
      live: false,
      fatal: true,
      reason: `live e2e preflight failed: ${url} returned HTTP ${res.status}: ${(await res.text()).slice(0, 240)}`,
    };
  } catch {
    return { live: false, fatal: false, reason: UNREACHABLE_REASON };
  } finally {
    clearTimeout(timeout);
  }
}
