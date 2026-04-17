import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
    getEnvConfig,
    getFidelityFormPath,
} from '../fixtures';
import { Split } from '../../src/types';

// Global test constants
const TEST_TIMEOUT = 180000;

const DEFAULT_SUBDOCUMENTS = [
    { name: 'form_section', description: 'Form sections with input fields, checkboxes, or signature areas' },
    { name: 'instructions', description: 'Instructions, terms and conditions, or explanatory text' },
    { name: 'header', description: 'Header sections with logos, titles, or document identifiers' },
];

function validateSplit(s: Split | null): void {
    expect(s).not.toBeNull();
    expect(s).toBeDefined();
    if (!s) return;

    expect(s.id).toBeDefined();
    expect(typeof s.id).toBe('string');
    expect(s.id.length).toBeGreaterThan(0);

    expect(s.file).toBeDefined();
    expect(s.model).toBeDefined();

    expect(Array.isArray(s.subdocuments)).toBe(true);
    expect(Array.isArray(s.output)).toBe(true);
    expect(s.output.length).toBeGreaterThan(0);

    s.output.forEach((entry) => {
        expect(typeof entry.name).toBe('string');
        expect(Array.isArray(entry.pages)).toBe(true);
    });
}

describe('Retab SDK Splits Resource Tests', () => {
    let client: Retab;
    let fidelityFormPath: string;

    beforeAll(() => {
        const envConfig = getEnvConfig();
        client = new Retab({
            apiKey: envConfig.retabApiKey,
            baseUrl: envConfig.retabApiBaseUrl,
        });

        fidelityFormPath = getFidelityFormPath();
    });

    describe('Splits CRUD roundtrip', () => {
        test('test_splits_create_get_delete', async () => {
            const split = await client.splits.create({
                document: fidelityFormPath,
                model: "retab-micro",
                subdocuments: DEFAULT_SUBDOCUMENTS,
            });
            validateSplit(split);

            const fetched = await client.splits.get(split.id);
            expect(fetched.id).toBe(split.id);

            const listed = await client.splits.list({ limit: 10 });
            expect(listed).toHaveProperty('data');
            expect(Array.isArray(listed.data)).toBe(true);

            await client.splits.delete(split.id);
        }, { timeout: TEST_TIMEOUT });
    });
});
