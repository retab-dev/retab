import { expect, test } from 'bun:test';
import { RetabBase } from '../../src/common/retab-base.js';

const document = (bytes: number) => ({
  filename: 'large.pdf',
  url: 'data:application/pdf;base64,' + Buffer.alloc(bytes, 65).toString('base64'),
});

function uploadClient(failUpload = false, failSubmission = false) {
  const calls: Array<{ url: URL; init: RequestInit }> = [];
  const client = new RetabBase({
    apiKey: 'test_key',
    fetch: (async (input, init = {}) => {
      const url = new URL(String(input));
      calls.push({ url, init });
      const json = (value: unknown, status = 200) =>
        new Response(JSON.stringify(value), {
          status,
          headers: { 'content-type': 'application/json' },
        });
      if (url.hostname === 'upload.example') {
        expect(new Headers(init.headers).has('Authorization')).toBe(false);
        return new Response(null, { status: failUpload ? 503 : 200 });
      }
      expect(new Headers(init.headers).get('Authorization')).toBe('Bearer test_key');
      if (url.pathname === '/v1/files/upload') {
        const body = JSON.parse(String(init.body));
        expect(body.content_type).toBe('application/pdf');
        return json({
          fileId: 'file_large',
          uploadUrl: 'https://upload.example/file',
          uploadMethod: 'PUT',
        });
      }
      if (url.pathname === '/v1/files/upload/file_large/complete') {
        return json({ filename: 'large.pdf', url: 'https://storage.retab.com/file_large' });
      }
      expect(Buffer.byteLength(String(init.body))).toBeLessThan(32 * 1024 * 1024);
      return json({ id: 'run_1' }, failSubmission ? 500 : 200);
    }) as typeof fetch,
  });
  return { client, calls };
}

test('uploads a large document before workflow submission without mutating caller data', async () => {
  const original = document(25 * 1024 * 1024);
  const body = { workflow_id: 'wf_1', documents: { start: original } };
  const { client, calls } = uploadClient();
  await client.request({ method: 'POST', path: '/v1/workflows/runs', body });
  expect(calls.map((call) => call.url.pathname)).toEqual([
    '/v1/files/upload',
    '/file',
    '/v1/files/upload/file_large/complete',
    '/v1/workflows/runs',
  ]);
  expect((calls[1].init.body as Uint8Array).byteLength).toBe(25 * 1024 * 1024);
  expect(JSON.parse(String(calls[3].init.body)).documents.start.url).toBe(
    'https://storage.retab.com/file_large'
  );
  expect(body.documents.start.url).toBe(original.url);
});

test('uses the combined request budget for multiple smaller documents', async () => {
  const { client, calls } = uploadClient();
  await client.request({
    method: 'POST',
    path: '/v1/workflows/runs',
    body: {
      workflow_id: 'wf_1',
      documents: { one: document(12 * 1024 * 1024), two: document(12 * 1024 * 1024) },
    },
  });
  expect(calls.filter((call) => call.url.hostname === 'upload.example')).toHaveLength(2);
});

test('uploads schema document arrays using the combined budget and preserves their shape', async () => {
  const first = document(15 * 1024 * 1024);
  const remote = { filename: 'remote.pdf', url: 'https://example.com/remote.pdf' };
  const last = document(15 * 1024 * 1024);
  const body = { documents: [first, remote, last], instructions: 'Keep array order' };
  const { client, calls } = uploadClient();
  await client.request({ method: 'POST', path: '/v1/schemas/generate', body });
  const submitted = JSON.parse(String(calls.at(-1)!.init.body));
  expect(calls.filter((call) => call.url.hostname === 'upload.example')).toHaveLength(2);
  expect(Array.isArray(submitted.documents)).toBe(true);
  expect(submitted.documents).toEqual([
    { filename: 'large.pdf', url: 'https://storage.retab.com/file_large' },
    remote,
    { filename: 'large.pdf', url: 'https://storage.retab.com/file_large' },
  ]);
  expect(submitted.instructions).toBe(body.instructions);
  expect(body.documents).toEqual([first, remote, last]);
  expect(first.url.startsWith('data:')).toBe(true);
  expect(last.url.startsWith('data:')).toBe(true);
});

test('keeps small documents and unrelated JSON data URLs unchanged', async () => {
  const body = { document: document(64), json_inputs: { source: document(64) } };
  const { client, calls } = uploadClient();
  await client.request({ method: 'POST', path: '/v1/parses', body });
  expect(calls).toHaveLength(1);
  expect(JSON.parse(String(calls[0].init.body))).toEqual(body);
});

test('does not finalize or submit after a failed storage upload', async () => {
  const { client, calls } = uploadClient(true);
  await expect(
    client.request({
      method: 'POST',
      path: '/v1/workflows/runs',
      body: {
        documents: { start: document(21 * 1024 * 1024) },
      },
    })
  ).rejects.toThrow('direct file upload failed');
  expect(calls).toHaveLength(2);
});
