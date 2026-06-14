// Offline unit tests for src/runtime/mime.ts coerceMimeData + helpers.
// Mirrors the Python SDK tests test_mime_data_urls.py / test_file_ref_contract.py.
// Pure-function tests — no network, no fetch injection required here.

import { describe, expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as os from 'node:os';
import * as path from 'node:path';
import { Readable } from 'node:stream';

import {
  coerceMimeData,
  retabStorageFileIdFromUrl,
  type FileRefDocumentInput,
  type MIMEData,
  type FileRefDocumentWire,
} from '../../src/runtime/mime.js';

// Minimal valid 1x1 PNG bytes (magic header 0x89 0x50 0x4e 0x47).
const PNG_BYTES = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x01]);
const PDF_BYTES = Buffer.from('%PDF-1.4\n%%EOF\n');
const JPEG_BYTES = Buffer.from([0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10]);

describe('coerceMimeData data-URL inputs', () => {
  test('decodes a data: URL string into MIMEData with synthetic filename', async () => {
    const dataUrl = `data:image/png;base64,${PNG_BYTES.toString('base64')}`;
    const result = (await coerceMimeData(dataUrl)) as MIMEData;
    expect(result.url).toBe(dataUrl);
    expect(result.filename).toBe('uploaded_file.png');
  });

  test('data: URL without subtype falls back to octet-stream -> bin', async () => {
    const dataUrl = 'data:,SGVsbG8='; // no media type
    const result = (await coerceMimeData(dataUrl)) as MIMEData;
    expect(result.url).toBe(dataUrl);
    // mime defaults to 'application/octet-stream', ext = 'octet-stream'
    expect(result.filename.startsWith('uploaded_file.')).toBe(true);
  });

  test('pre-shaped MIMEData object passes through unchanged', async () => {
    const shaped: MIMEData = { filename: 'doc.pdf', url: 'data:application/pdf;base64,JVBERi0=' };
    const result = (await coerceMimeData(shaped)) as MIMEData;
    expect(result).toBe(shaped);
  });
});

describe('coerceMimeData buffer / base64 / stream coercion', () => {
  test('Buffer is encoded as a data: URL with detected mime', async () => {
    const result = (await coerceMimeData(PNG_BYTES)) as MIMEData;
    expect(result.filename).toBe('uploaded_file.png');
    expect(result.url).toBe(`data:image/png;base64,${PNG_BYTES.toString('base64')}`);
  });

  test('PDF Buffer is sniffed to application/pdf', async () => {
    const result = (await coerceMimeData(PDF_BYTES)) as MIMEData;
    expect(result.url.startsWith('data:application/pdf;base64,')).toBe(true);
    expect(result.filename).toBe('uploaded_file.pdf');
  });

  test('base64 string (non-path, non-url) is decoded and re-encoded', async () => {
    const b64 = PNG_BYTES.toString('base64');
    const result = (await coerceMimeData(b64)) as MIMEData;
    expect(result.filename).toBe('uploaded_file.bin');
    // the bytes round-trip through magic-byte detection back to image/png
    expect(result.url).toBe(`data:image/png;base64,${b64}`);
  });

  test('filesystem path is read and base64-encoded with basename filename', async () => {
    const tmp = path.join(os.tmpdir(), `retab-mime-test-${Date.now()}.png`);
    fs.writeFileSync(tmp, PNG_BYTES);
    try {
      const result = (await coerceMimeData(tmp)) as MIMEData;
      expect(result.filename).toBe(path.basename(tmp));
      expect(result.url).toBe(`data:image/png;base64,${PNG_BYTES.toString('base64')}`);
    } finally {
      fs.unlinkSync(tmp);
    }
  });

  test('Readable stream is buffered and encoded; path-less stream gets synthetic filename', async () => {
    const stream = Readable.from([JPEG_BYTES]);
    const result = (await coerceMimeData(stream)) as MIMEData;
    expect(result.filename).toBe('uploaded_file.jpeg');
    expect(result.url).toBe(`data:image/jpeg;base64,${JPEG_BYTES.toString('base64')}`);
  });

  test('Readable stream with a .path uses its basename as filename', async () => {
    const stream = Readable.from([PDF_BYTES]) as Readable & { path?: string };
    stream.path = '/some/dir/report.pdf';
    const result = (await coerceMimeData(stream)) as MIMEData;
    expect(result.filename).toBe('report.pdf');
    expect(result.url.startsWith('data:application/pdf;base64,')).toBe(true);
  });
});

describe('coerceMimeData FileRefDocumentInput camelCase -> wire', () => {
  test('mimeType (camelCase) is mapped to mime_type on the wire', async () => {
    const input: FileRefDocumentInput = {
      id: 'file_abc',
      filename: 'stored.pdf',
      mimeType: 'application/pdf',
    };
    const result = (await coerceMimeData(input)) as FileRefDocumentWire;
    expect(result).toEqual({
      id: 'file_abc',
      filename: 'stored.pdf',
      mime_type: 'application/pdf',
    });
    expect('mimeType' in result).toBe(false);
  });

  test('snake_case mime_type input is preserved on the wire', async () => {
    const input: FileRefDocumentInput = {
      id: 'file_xyz',
      filename: 'stored.png',
      mime_type: 'image/png',
    };
    const result = (await coerceMimeData(input)) as FileRefDocumentWire;
    expect(result.mime_type).toBe('image/png');
    expect(result.id).toBe('file_xyz');
  });

  test('snake_case wins over camelCase when both are present', async () => {
    const input: FileRefDocumentInput = {
      id: 'file_dup',
      filename: 'stored.bin',
      mime_type: 'application/json',
      mimeType: 'text/plain',
    };
    const result = (await coerceMimeData(input)) as FileRefDocumentWire;
    expect(result.mime_type).toBe('application/json');
  });
});

describe('coerceMimeData https:// passthrough + filename derivation', () => {
  test('https URL passthrough derives filename from last path segment', async () => {
    const result = (await coerceMimeData('https://example.com/files/remote.pdf?download=1')) as MIMEData;
    expect(result.url).toBe('https://example.com/files/remote.pdf?download=1');
    expect(result.filename).toBe('remote.pdf');
  });

  test('https URL with trailing slash falls back to remote_file', async () => {
    const result = (await coerceMimeData('https://example.com/')) as MIMEData;
    expect(result.filename).toBe('remote_file');
    expect(result.url).toBe('https://example.com/');
  });

  test('URL instance is stringified then passthrough', async () => {
    const result = (await coerceMimeData(new URL('https://cdn.example.com/a/b/photo.jpg'))) as MIMEData;
    expect(result.filename).toBe('photo.jpg');
    expect(result.url).toBe('https://cdn.example.com/a/b/photo.jpg');
  });

  test('FileRefInput {type: file_url} passes through the URL', async () => {
    const result = (await coerceMimeData({
      type: 'file_url',
      url: 'https://example.com/data/invoice.pdf',
    })) as MIMEData;
    expect(result.filename).toBe('invoice.pdf');
    expect(result.url).toBe('https://example.com/data/invoice.pdf');
  });
});

describe('retabStorageFileIdFromUrl parsing', () => {
  test('parses the file id from a clean storage URL', () => {
    expect(retabStorageFileIdFromUrl('https://storage.retab.com/org_1/file_123.pdf')).toBe(
      'file_123'
    );
  });

  test('rejects non-https schemes', () => {
    expect(retabStorageFileIdFromUrl('http://storage.retab.com/org_1/file_123.pdf')).toBeUndefined();
  });

  test('rejects a non-storage host', () => {
    expect(retabStorageFileIdFromUrl('https://example.com/org_1/file_123.pdf')).toBeUndefined();
  });

  test('rejects URLs carrying a query string or hash', () => {
    expect(
      retabStorageFileIdFromUrl('https://storage.retab.com/org_1/file_123.pdf?x=1')
    ).toBeUndefined();
    expect(
      retabStorageFileIdFromUrl('https://storage.retab.com/org_1/file_123.pdf#frag')
    ).toBeUndefined();
  });

  test('rejects wrong path-part counts', () => {
    expect(retabStorageFileIdFromUrl('https://storage.retab.com/file_123.pdf')).toBeUndefined();
    expect(
      retabStorageFileIdFromUrl('https://storage.retab.com/a/b/file_123.pdf')
    ).toBeUndefined();
  });

  test('rejects names with no extension or a leading/trailing dot', () => {
    expect(retabStorageFileIdFromUrl('https://storage.retab.com/org_1/file_123')).toBeUndefined();
    expect(retabStorageFileIdFromUrl('https://storage.retab.com/org_1/.pdf')).toBeUndefined();
    expect(retabStorageFileIdFromUrl('https://storage.retab.com/org_1/file_123.')).toBeUndefined();
  });

  test('returns undefined for an unparseable URL', () => {
    expect(retabStorageFileIdFromUrl('not a url at all')).toBeUndefined();
  });
});
