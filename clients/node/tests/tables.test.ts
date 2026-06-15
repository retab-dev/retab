import { describe, expect, test } from 'bun:test';

import { Tables } from '../src/tables/tables.js';

describe('Tables.download', () => {
  test('returns a Blob without text-decoding the response', async () => {
    const client = {
      request: async () => {
        throw new Error('download must use the raw Blob request path');
      },
      requestBlob: async ({ path }: { path: string }) => {
        expect(path).toBe('/v1/tables/tbl_test/download');
        return new Blob([new Uint8Array([0x6e, 0x61, 0x6d, 0x65])], { type: 'text/csv' });
      },
    };

    const resource = new Tables(client as never);
    const downloaded = await resource.download('tbl_test');

    expect(downloaded).toBeInstanceOf(Blob);
    expect(await downloaded.text()).toBe('name');
  });
});
