// Offline unit tests for src/common/exceptions.ts.
// Covers parseRetabError status->class mapping directly, and end-to-end by
// driving RetabBase.request through an injected fetch returning non-ok
// Responses (the documented fetch-injection pattern).

import { describe, expect, test } from 'bun:test';

import {
  RetabError,
  RetabAuthenticationError,
  RetabPermissionError,
  RetabNotFoundError,
  RetabRateLimitError,
  parseRetabError,
} from '../../src/common/exceptions.js';
import { RetabBase } from '../../src/common/retab-base.js';

describe('parseRetabError status mapping (direct)', () => {
  test('401 -> RetabAuthenticationError', () => {
    const err = parseRetabError(401, 'unauthorized');
    expect(err).toBeInstanceOf(RetabAuthenticationError);
    expect(err).toBeInstanceOf(RetabError);
    expect(err.status).toBe(401);
    expect(err.responseBody).toBe('unauthorized');
    expect(err.name).toBe('RetabAuthenticationError');
  });

  test('403 -> RetabPermissionError', () => {
    const err = parseRetabError(403, 'forbidden');
    expect(err).toBeInstanceOf(RetabPermissionError);
    expect(err.status).toBe(403);
    expect(err.responseBody).toBe('forbidden');
    expect(err.name).toBe('RetabPermissionError');
  });

  test('404 -> RetabNotFoundError', () => {
    const err = parseRetabError(404, 'missing');
    expect(err).toBeInstanceOf(RetabNotFoundError);
    expect(err.status).toBe(404);
    expect(err.responseBody).toBe('missing');
    expect(err.name).toBe('RetabNotFoundError');
  });

  test('429 -> RetabRateLimitError', () => {
    const err = parseRetabError(429, 'slow down');
    expect(err).toBeInstanceOf(RetabRateLimitError);
    expect(err.status).toBe(429);
    expect(err.responseBody).toBe('slow down');
    expect(err.name).toBe('RetabRateLimitError');
  });

  test('500 (and other statuses) -> base RetabError', () => {
    const err = parseRetabError(500, 'boom');
    expect(err).toBeInstanceOf(RetabError);
    expect(err).not.toBeInstanceOf(RetabAuthenticationError);
    expect(err).not.toBeInstanceOf(RetabPermissionError);
    expect(err).not.toBeInstanceOf(RetabNotFoundError);
    expect(err).not.toBeInstanceOf(RetabRateLimitError);
    expect(err.status).toBe(500);
    expect(err.responseBody).toBe('boom');
    expect(err.message).toBe('Retab API error (500)');
    expect(err.name).toBe('RetabError');
  });

  test('418 (unmapped 4xx) -> base RetabError', () => {
    const err = parseRetabError(418, 'teapot');
    expect(err.constructor).toBe(RetabError);
    expect(err.status).toBe(418);
  });
});

// Expose the protected-by-convention `request` on a tiny subclass so the test
// can drive a single HTTP round-trip through the injected fetch.
class Probe extends RetabBase {
  call<T>(opts: Parameters<RetabBase['request']>[0]): Promise<T> {
    return this.request<T>(opts);
  }
}

function probeWithStatus(status: number, body: string): Probe {
  return new Probe({
    apiKey: 'test_key',
    fetch: async () =>
      new Response(body, { status, headers: { 'content-type': 'application/json' } }),
  });
}

describe('RetabBase.request surfaces parseRetabError for non-ok responses', () => {
  const cases: Array<[number, new (...args: never[]) => RetabError]> = [
    [401, RetabAuthenticationError],
    [403, RetabPermissionError],
    [404, RetabNotFoundError],
    [429, RetabRateLimitError],
    [500, RetabError],
  ];

  for (const [status, ctor] of cases) {
    test(`status ${status} throws ${ctor.name} carrying status + body`, async () => {
      const body = `{"detail":"failure ${status}"}`;
      const probe = probeWithStatus(status, body);
      let thrown: unknown;
      try {
        await probe.call({ method: 'GET', path: '/v1/anything' });
      } catch (e) {
        thrown = e;
      }
      expect(thrown).toBeInstanceOf(ctor);
      expect(thrown).toBeInstanceOf(RetabError);
      const err = thrown as RetabError;
      expect(err.status).toBe(status);
      expect(err.responseBody).toBe(body);
    });
  }

  test('error body is read verbatim (not parsed) and preserved on responseBody', async () => {
    const probe = probeWithStatus(403, 'plain text rejection');
    await expect(probe.call({ method: 'POST', path: '/v1/x', body: { a: 1 } })).rejects.toMatchObject(
      {
        status: 403,
        responseBody: 'plain text rejection',
      }
    );
  });
});
