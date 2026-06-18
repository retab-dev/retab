import { describe, expect, test } from 'bun:test';

import { probeLiveServer } from './live-preflight';

describe('live e2e preflight', () => {
  test('skips when credentials are missing', async () => {
    const result = await probeLiveServer(undefined, 'http://localhost:4000', async () => {
      throw new Error('fetch should not be called');
    });

    expect(result.live).toBe(false);
    if (!result.live) {
      expect(result.fatal).toBe(false);
      expect(result.reason).toContain('set RETAB_API_KEY');
    }
  });

  test('accepts an authenticated response', async () => {
    const seen: string[] = [];
    const result = await probeLiveServer(
      'sk_valid',
      'http://localhost:4000/',
      async (url, init) => {
        seen.push(url);
        expect(init.headers).toEqual({
          Authorization: 'Bearer sk_valid',
          Accept: 'application/json',
        });
        return new Response('{}', { status: 200 });
      }
    );

    expect(result).toEqual({ live: true, reason: '' });
    expect(seen).toEqual(['http://localhost:4000/v1/files?limit=1']);
  });

  test('fails when the configured key is rejected', async () => {
    const result = await probeLiveServer('sk_bad', 'http://localhost:4000', async () => {
      return new Response('invalid key', { status: 401 });
    });

    expect(result.live).toBe(false);
    if (!result.live) {
      expect(result.fatal).toBe(true);
      expect(result.reason).toContain('RETAB_API_KEY');
    }
  });

  test('skips when the server is unreachable', async () => {
    const result = await probeLiveServer('sk_valid', 'http://localhost:4000', async () => {
      throw new Error('connection refused');
    });

    expect(result.live).toBe(false);
    if (!result.live) {
      expect(result.fatal).toBe(false);
      expect(result.reason).toContain('unreachable');
    }
  });

  test('fails on unexpected non-auth HTTP responses', async () => {
    const result = await probeLiveServer('sk_valid', 'http://localhost:4000', async () => {
      return new Response('missing route details', { status: 404 });
    });

    expect(result.live).toBe(false);
    if (!result.live) {
      expect(result.fatal).toBe(true);
      expect(result.reason).toContain('HTTP 404');
      expect(result.reason).toContain('missing route details');
    }
  });
});
