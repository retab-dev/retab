import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
  getEnvConfig,
  getBookingConfirmationFilePath1,
  getBookingConfirmationJsonSchema,
} from '../fixtures';
import { ExtractionV2 } from '../../src/types';

const TEST_TIMEOUT = 180000;

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
});
