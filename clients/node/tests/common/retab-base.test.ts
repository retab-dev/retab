// Offline unit tests for src/common/retab-base.ts.
// Drives RetabBase.request through an injected fetch, capturing the outgoing
// Request to assert query serialization, header injection, body framing, the
// /v1 baseUrl strip warning, and response decoding.

import { afterEach, describe, expect, mock, test } from 'bun:test';

import { RetabBase, type RequestOptions } from '../../src/common/retab-base.js';

class Probe extends RetabBase {
  call<T>(opts: RequestOptions): Promise<T> {
    return this.request<T>(opts);
  }
}

type Captured = { url: URL; init: RequestInit };

function probeCapturing(
  response: () => Response,
  opts?: { baseUrl?: string }
): { probe: Probe; captured: () => Captured } {
  let captured: Captured | undefined;
  const probe = new Probe({
    apiKey: 'secret_key',
    baseUrl: opts?.baseUrl,
    fetch: async (input, init) => {
      captured = { url: new URL(String(input)), init: init ?? {} };
      return response();
    },
  });
  return {
    probe,
    captured: () => {
      if (!captured) throw new Error('fetch was not invoked');
      return captured;
    },
  };
}

function jsonResponse(payload: unknown, status = 200): Response {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'content-type': 'application/json' },
  });
}

describe('RetabBase.request query serialization', () => {
  test('array query params are repeated, not CSV-joined', async () => {
    const { probe, captured } = probeCapturing(() => jsonResponse({ ok: true }));
    await probe.call({
      method: 'GET',
      path: '/v1/items',
      query: { tag: ['a', 'b', 'c'] },
    });
    expect(captured().url.searchParams.getAll('tag')).toEqual(['a', 'b', 'c']);
    // Must NOT collapse to a single CSV value.
    expect(captured().url.searchParams.get('tag')).toBe('a');
    expect(captured().url.search).toContain('tag=a&tag=b&tag=c');
  });

  test('Date query values are serialized via toISOString', async () => {
    const { probe, captured } = probeCapturing(() => jsonResponse({}));
    const when = new Date('2026-01-02T03:04:05.678Z');
    await probe.call({ method: 'GET', path: '/v1/items', query: { since: when } });
    expect(captured().url.searchParams.get('since')).toBe('2026-01-02T03:04:05.678Z');
  });

  test('null and undefined query values are omitted entirely', async () => {
    const { probe, captured } = probeCapturing(() => jsonResponse({}));
    await probe.call({
      method: 'GET',
      path: '/v1/items',
      query: { keep: 'yes', dropNull: null, dropUndef: undefined },
    });
    const sp = captured().url.searchParams;
    expect(sp.get('keep')).toBe('yes');
    expect(sp.has('dropNull')).toBe(false);
    expect(sp.has('dropUndef')).toBe(false);
  });

  test('null/undefined entries inside an array query are skipped', async () => {
    const { probe, captured } = probeCapturing(() => jsonResponse({}));
    await probe.call({
      method: 'GET',
      path: '/v1/items',
      query: { tag: ['a', null, undefined, 'b'] },
    });
    expect(captured().url.searchParams.getAll('tag')).toEqual(['a', 'b']);
  });
});

describe('RetabBase.request headers and body framing', () => {
  test('Authorization header is injected from apiKey', async () => {
    const { probe, captured } = probeCapturing(() => jsonResponse({}));
    await probe.call({ method: 'GET', path: '/v1/items' });
    const headers = captured().init.headers as Record<string, string>;
    expect(headers['Authorization']).toBe('Bearer secret_key');
    expect(headers['Accept']).toBe('application/json');
  });

  test('Content-Type set + JSON body emitted only when a body is present', async () => {
    const { probe, captured } = probeCapturing(() => jsonResponse({}));
    await probe.call({ method: 'POST', path: '/v1/items', body: { name: 'x' } });
    const headers = captured().init.headers as Record<string, string>;
    expect(headers['Content-Type']).toBe('application/json');
    expect(captured().init.body).toBe(JSON.stringify({ name: 'x' }));
    expect(captured().init.method).toBe('POST');
  });

  test('no Content-Type and no body for body-less requests', async () => {
    const { probe, captured } = probeCapturing(() => jsonResponse({}));
    await probe.call({ method: 'GET', path: '/v1/items' });
    const headers = captured().init.headers as Record<string, string>;
    expect('Content-Type' in headers).toBe(false);
    expect(captured().init.body).toBeUndefined();
  });

  test('caller-supplied headers are merged in', async () => {
    const { probe, captured } = probeCapturing(() => jsonResponse({}));
    await probe.call({
      method: 'GET',
      path: '/v1/items',
      headers: { 'X-Trace': 'abc' },
    });
    const headers = captured().init.headers as Record<string, string>;
    expect(headers['X-Trace']).toBe('abc');
    expect(headers['Authorization']).toBe('Bearer secret_key');
  });
});

describe('RetabBase trailing /v1 baseUrl strip', () => {
  afterEach(() => {
    mock.restore();
  });

  test('a baseUrl ending in /v1 is stripped and emits console.warn', async () => {
    const warn = mock(() => {});
    const original = console.warn;
    console.warn = warn;
    try {
      const { probe, captured } = probeCapturing(() => jsonResponse({}), {
        baseUrl: 'https://api.example.com/v1',
      });
      await probe.call({ method: 'GET', path: '/v1/items' });
      // Path keeps /v1; base no longer carries it -> single /v1, not /v1/v1.
      expect(captured().url.pathname).toBe('/v1/items');
      expect(warn).toHaveBeenCalledTimes(1);
      expect(String(warn.mock.calls[0][0])).toContain('ends with a version segment');
    } finally {
      console.warn = original;
    }
  });

  test('a baseUrl ending in /v1/ is also stripped', async () => {
    const warn = mock(() => {});
    const original = console.warn;
    console.warn = warn;
    try {
      const { probe, captured } = probeCapturing(() => jsonResponse({}), {
        baseUrl: 'https://api.example.com/v1/',
      });
      await probe.call({ method: 'GET', path: '/v1/items' });
      expect(captured().url.pathname).toBe('/v1/items');
      expect(warn).toHaveBeenCalledTimes(1);
    } finally {
      console.warn = original;
    }
  });

  test('a normal baseUrl does not warn', async () => {
    const warn = mock(() => {});
    const original = console.warn;
    console.warn = warn;
    try {
      const { probe, captured } = probeCapturing(() => jsonResponse({}), {
        baseUrl: 'https://api.example.com',
      });
      await probe.call({ method: 'GET', path: '/v1/items' });
      expect(captured().url.pathname).toBe('/v1/items');
      expect(warn).not.toHaveBeenCalled();
    } finally {
      console.warn = original;
    }
  });
});

describe('RetabBase.request response decoding', () => {
  test('204 No Content resolves to undefined', async () => {
    const { probe } = probeCapturing(() => new Response(null, { status: 204 }));
    const result = await probe.call({ method: 'DELETE', path: '/v1/items/1' });
    expect(result).toBeUndefined();
  });

  test('application/json content-type is parsed as JSON', async () => {
    const { probe } = probeCapturing(() => jsonResponse({ id: 'abc', n: 7 }));
    const result = await probe.call<{ id: string; n: number }>({
      method: 'GET',
      path: '/v1/items/abc',
    });
    expect(result).toEqual({ id: 'abc', n: 7 });
  });

  test('non-JSON content-type returns the raw text body', async () => {
    const { probe } = probeCapturing(
      () => new Response('hello,world', { status: 200, headers: { 'content-type': 'text/csv' } })
    );
    const result = await probe.call<string>({ method: 'GET', path: '/v1/export' });
    expect(result).toBe('hello,world');
  });

  test('missing content-type falls back to text', async () => {
    const { probe } = probeCapturing(() => new Response('plain', { status: 200 }));
    const result = await probe.call<string>({ method: 'GET', path: '/v1/raw' });
    expect(result).toBe('plain');
  });
});

describe('RetabBase construction guards', () => {
  test('missing apiKey throws', () => {
    expect(() => new Probe({ apiKey: '' })).toThrow('Retab: apiKey is required');
  });
});
