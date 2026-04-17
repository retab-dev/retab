import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
    getEnvConfig,
    getBookingConfirmationFilePath1,
} from '../fixtures';
import { Classification } from '../../src/types';

// Global test constants
const TEST_TIMEOUT = 180000;

const DEFAULT_CATEGORIES = [
    { name: 'invoice', description: 'Invoice documents with billing information' },
    { name: 'receipt', description: 'Receipt documents for payments' },
    { name: 'booking', description: 'Booking confirmations or reservations' },
];

function validateClassification(c: Classification | null): void {
    expect(c).not.toBeNull();
    expect(c).toBeDefined();
    if (!c) return;

    expect(c.id).toBeDefined();
    expect(typeof c.id).toBe('string');
    expect(c.id.length).toBeGreaterThan(0);

    expect(c.file).toBeDefined();
    expect(c.model).toBeDefined();
    expect(Array.isArray(c.categories)).toBe(true);

    expect(c.output).toBeDefined();
    expect(typeof c.output.category).toBe('string');
    expect(typeof c.output.reasoning).toBe('string');

    expect(c.consensus).toBeDefined();
    expect(Array.isArray(c.consensus.choices)).toBe(true);
}

describe('Retab SDK Classifications Resource Tests', () => {
    let client: Retab;
    let bookingConfirmationFilePath1: string;

    beforeAll(() => {
        const envConfig = getEnvConfig();
        client = new Retab({
            apiKey: envConfig.retabApiKey,
            baseUrl: envConfig.retabApiBaseUrl,
        });

        bookingConfirmationFilePath1 = getBookingConfirmationFilePath1();
    });

    describe('Classifications CRUD roundtrip', () => {
        test('test_classifications_create_get_delete', async () => {
            const classification = await client.classifications.create({
                document: bookingConfirmationFilePath1,
                model: "retab-micro",
                categories: DEFAULT_CATEGORIES,
            });
            validateClassification(classification);

            const fetched = await client.classifications.get(classification.id);
            expect(fetched.id).toBe(classification.id);

            const listed = await client.classifications.list({ limit: 10 });
            expect(listed).toHaveProperty('data');
            expect(Array.isArray(listed.data)).toBe(true);

            await client.classifications.delete(classification.id);
        }, { timeout: TEST_TIMEOUT });
    });
});
