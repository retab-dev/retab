import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
    getEnvConfig,
    getFidelityFormPath,
    getFidelityInstructions,
} from '../fixtures';
import { Edit } from '../../src/types';

// Global test constants
const TEST_TIMEOUT = 240000;

function validateEdit(e: Edit | null): void {
    expect(e).not.toBeNull();
    expect(e).toBeDefined();
    if (!e) return;

    expect(e.id).toBeDefined();
    expect(typeof e.id).toBe('string');
    expect(e.id.length).toBeGreaterThan(0);

    expect(e.file).toBeDefined();
    expect(e.model).toBeDefined();
    expect(typeof e.instructions).toBe('string');
    expect(e.config).toBeDefined();

    expect(e.data).toBeDefined();
    expect(Array.isArray(e.data.form_data)).toBe(true);
    expect(e.data.filled_document).toBeDefined();
}

describe('Retab SDK Edits Resource Tests', () => {
    let client: Retab;
    let fidelityFormPath: string;
    let instructions: string;

    beforeAll(() => {
        const envConfig = getEnvConfig();
        client = new Retab({
            apiKey: envConfig.retabApiKey,
            baseUrl: envConfig.retabApiBaseUrl,
        });

        fidelityFormPath = getFidelityFormPath();
        instructions = getFidelityInstructions();
    });

    describe('Edits CRUD roundtrip', () => {
        test('test_edits_create_get_delete', async () => {
            const edit = await client.edits.create({
                document: fidelityFormPath,
                instructions,
                model: "retab-small",
            });
            validateEdit(edit);

            const fetched = await client.edits.get(edit.id);
            expect(fetched.id).toBe(edit.id);

            const listed = await client.edits.list({ limit: 10 });
            expect(listed).toHaveProperty('data');
            expect(Array.isArray(listed.data)).toBe(true);

            await client.edits.delete(edit.id);
        }, { timeout: TEST_TIMEOUT });
    });
});
