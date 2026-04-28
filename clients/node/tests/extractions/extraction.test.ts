import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import APIV1 from '../../src/api/client.js';
import { AbstractClient } from '../../src/client.js';
import {
  getEnvConfig,
  getBookingConfirmationFilePath1,
  getBookingConfirmationJsonSchema,
} from '../fixtures';
import { ExtractionV2 } from '../../src/types';

const TEST_TIMEOUT = 180000;

type RecordedRequest = {
  url: string;
  method: string;
  params?: Record<string, unknown>;
  headers?: Record<string, unknown>;
  bodyMime?: 'application/json' | 'multipart/form-data';
  body?: Record<string, unknown>;
};

class MockClient extends AbstractClient {
  public requests: RecordedRequest[] = [];

  protected async _fetch(params: RecordedRequest): Promise<Response> {
    this.requests.push(params);

    if (params.url === '/extractions/stream') {
      const chunks = [
        JSON.stringify({
          id: 'chatcmpl-1',
          object: 'chat.completion.chunk',
          created: 1710000000,
          model: 'retab-micro',
          choices: [
            {
              index: 0,
              finish_reason: null,
              delta: {
                role: 'assistant',
                content: '{"booking_reference":"BK-1"}',
                flat_parsed: { booking_reference: 'BK-1' },
                flat_likelihoods: { booking_reference: 1 },
                flat_deleted_keys: [],
                is_valid_json: true,
              },
            },
          ],
          extraction_id: 'extr_stream_123',
        }),
        JSON.stringify({
          id: 'chatcmpl-2',
          object: 'chat.completion.chunk',
          created: 1710000001,
          model: 'retab-micro',
          choices: [
            {
              index: 0,
              finish_reason: 'stop',
              delta: {
                content: '',
                flat_parsed: {},
                flat_likelihoods: {},
                flat_deleted_keys: [],
                is_valid_json: true,
              },
            },
          ],
          extraction_id: 'extr_stream_123',
        }),
      ].join('');

      return new Response(chunks, {
        status: 200,
        headers: { 'Content-Type': 'application/stream+json' },
      });
    }

    return new Response(
      JSON.stringify({
        id: 'extr_123',
        file: {
          id: 'file_123',
          filename: 'booking_confirmation_1.jpg',
          mime_type: 'image/jpeg',
        },
        model: 'retab-micro',
        json_schema: { type: 'object' },
        n_consensus: 1,
        image_resolution_dpi: 192,
        output: { booking_reference: 'BK-1' },
        consensus: { choices: [], likelihoods: null },
        metadata: {},
      }),
      {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  }
}

function validateExtraction(extraction: ExtractionV2 | null): void {
  expect(extraction).not.toBeNull();
  expect(extraction).toBeDefined();
  if (!extraction) return;

  expect(extraction.id).toBeDefined();
  expect(typeof extraction.id).toBe('string');
  expect(extraction.id.length).toBeGreaterThan(0);

  expect(extraction.file).toBeDefined();
  expect(extraction.model).toBeDefined();
  expect(extraction.json_schema).toBeDefined();
  expect(extraction.output).toBeDefined();
  expect(extraction.consensus).toBeDefined();
  expect(Array.isArray(extraction.consensus.choices)).toBe(true);
}

describe('Retab SDK Extractions Resource Tests', () => {
  let client: Retab;
  let bookingConfirmationFilePath1: string;
  let bookingConfirmationJsonSchema: Record<string, any>;

  beforeAll(() => {
    const envConfig = getEnvConfig();
    client = new Retab({
      apiKey: envConfig.retabApiKey,
      baseUrl: envConfig.retabApiBaseUrl,
    });

    bookingConfirmationFilePath1 = getBookingConfirmationFilePath1();
    bookingConfirmationJsonSchema = getBookingConfirmationJsonSchema();
  });

  describe('Extractions CRUD roundtrip', () => {
    test(
      'test_extractions_create_get_list_delete',
      async () => {
        const extraction = await client.extractions.create({
          document: bookingConfirmationFilePath1,
          json_schema: bookingConfirmationJsonSchema,
          model: 'retab-micro',
        });
        validateExtraction(extraction);

        const fetched = await client.extractions.get(extraction.id);
        expect(fetched.id).toBe(extraction.id);
        expect(fetched.output).toEqual(extraction.output);

        const listed = await client.extractions.list({ limit: 10 });
        expect(listed).toHaveProperty('data');
        expect(Array.isArray(listed.data)).toBe(true);

        await client.extractions.delete(extraction.id);
      },
      { timeout: TEST_TIMEOUT }
    );
  });

  describe('Extractions streaming request tests', () => {
    test('createStream posts to the modern streaming route with stream body fields', async () => {
      const mockClient = new MockClient();
      const api = new APIV1(mockClient);
      const additionalMessages = [
        { role: 'developer', content: 'Use booking fields only.' },
        { role: 'user', content: 'Extract the booking reference.' },
      ];

      const stream = await api.extractions.createStream(
        {
          document: bookingConfirmationFilePath1,
          json_schema: bookingConfirmationJsonSchema,
          model: 'retab-micro',
          image_resolution_dpi: 183,
          n_consensus: 3,
          instructions: 'Return only valid structured data.',
          metadata: { case: 'stream' },
          additional_messages: additionalMessages,
          bust_cache: true,
        },
        {
          body: {
            chunking_keys: { line_items: 'items' },
          },
          params: { debug: '1' },
          headers: { 'Idempotency-Key': 'idem_1' },
        }
      );

      const chunks = [];
      for await (const chunk of stream) {
        chunks.push(chunk);
      }

      const request = mockClient.requests[0];
      expect(request?.url).toBe('/extractions/stream');
      expect(request?.method).toBe('POST');
      expect(request?.params).toEqual({ debug: '1' });
      expect(request?.headers).toEqual({ 'Idempotency-Key': 'idem_1' });
      expect(request?.body?.stream).toBe(true);
      expect((request?.body?.document as { filename?: string }).filename).toBe(
        'booking_confirmation_1.jpg'
      );
      expect(request?.body?.json_schema).toEqual(bookingConfirmationJsonSchema);
      expect(request?.body?.model).toBe('retab-micro');
      expect(request?.body?.image_resolution_dpi).toBe(183);
      expect(request?.body?.n_consensus).toBe(3);
      expect(request?.body?.instructions).toBe('Return only valid structured data.');
      expect(request?.body?.metadata).toEqual({ case: 'stream' });
      expect(request?.body?.additional_messages).toEqual(additionalMessages);
      expect(request?.body?.bust_cache).toBe(true);
      expect(request?.body?.chunking_keys).toEqual({ line_items: 'items' });

      expect(chunks).toHaveLength(2);
      expect(chunks[0]?.extraction_id).toBe('extr_stream_123');
      expect(chunks[0]?.choices[0]?.delta?.flat_parsed).toEqual({
        booking_reference: 'BK-1',
      });
      expect(chunks[0]?.choices[0]?.delta?.flat_likelihoods).toEqual({
        booking_reference: 1,
      });
      expect(chunks[1]?.choices[0]?.finish_reason).toBe('stop');
    });

    test('prepare_createStream captures the streaming request without performing a network call', async () => {
      const api = new APIV1(new MockClient());

      const request = await api.extractions.prepare_createStream({
        document: bookingConfirmationFilePath1,
        json_schema: bookingConfirmationJsonSchema,
        model: 'retab-micro',
      });

      expect(request.url).toBe('/extractions/stream');
      expect(request.method).toBe('POST');
      expect(request.body?.stream).toBe(true);
      expect((request.body?.document as { filename?: string }).filename).toBe(
        'booking_confirmation_1.jpg'
      );
    });
  });
});
