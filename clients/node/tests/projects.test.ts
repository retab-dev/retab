import { randomBytes } from 'crypto';
// @ts-ignore - bun test types may not be available in lint environment
import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../src/index.js';
import {
    getEnvConfig,
    getBookingConfirmationJsonSchema,
    getBookingConfirmationFilePath1,
    getBookingConfirmationFilePath2,
    getFidelityFormPath,
} from './fixtures';

// Simple ID generator to replace nanoid
function generateId(): string {
    return randomBytes(6).toString('hex');
}

// Global test constants
const TEST_TIMEOUT = 180000;

describe('Retab SDK Tests', () => {
    let client: Retab;
    let bookingConfirmationJsonSchema: Record<string, any>;
    let bookingConfirmationFilePath1: string;
    let bookingConfirmationFilePath2: string;

    beforeAll(() => {
        const envConfig = getEnvConfig();
        client = new Retab({
            apiKey: envConfig.retabApiKey,
            baseUrl: envConfig.retabApiBaseUrl,
        });

        bookingConfirmationJsonSchema = getBookingConfirmationJsonSchema();
        bookingConfirmationFilePath1 = getBookingConfirmationFilePath1();
        bookingConfirmationFilePath2 = getBookingConfirmationFilePath2();
    });

    describe('Authentication and Basic SDK Functionality', () => {
        test('test_authentication_validation', async () => {
            // Store original environment variable
            const originalEnvKey = process.env.RETAB_API_KEY;
            const envConfig = getEnvConfig(); // Get config before manipulating env

            try {
                // Test 1: Initialize without API key (should throw error)
                delete process.env.RETAB_API_KEY; // Remove env variable temporarily

                expect(() => {
                    new Retab();
                }).toThrow('Authentication required: Please provide an API key');

                // Test 2: Initialize with API key in constructor
                const clientWithApiKey = new Retab({ apiKey: envConfig.retabApiKey });
                expect(clientWithApiKey).toBeDefined();

                // Test 3: Initialize with environment variable
                process.env.RETAB_API_KEY = envConfig.retabApiKey;

                const clientWithEnv = new Retab();
                expect(clientWithEnv).toBeDefined();

            } finally {
                // Clean up environment variable
                if (originalEnvKey) {
                    process.env.RETAB_API_KEY = originalEnvKey;
                } else {
                    delete process.env.RETAB_API_KEY;
                }
            }
        }, { timeout: TEST_TIMEOUT });

        test('test_projects_extract_without_iteration_id', async () => {
            const evaluationName = `test_extract_no_iter_${generateId()}`;

            const evaluation = await client.evals.extract.create({
                name: evaluationName,
                json_schema: bookingConfirmationJsonSchema,
            });
            const evalId = evaluation.id;

            const publishedEvaluation = await client.evals.extract.publish(evalId);
            expect(publishedEvaluation.id).toBe(evalId);

            try {
                const response: any = await client.evals.extract.process({
                    eval_id: evalId,
                    document: bookingConfirmationFilePath1,
                });

                expect(response).toBeDefined();
                expect(response.id).toBeDefined();
                expect(response.choices).toBeDefined();
                expect(response.choices.length).toBeGreaterThan(0);
                expect(response.choices[0].message).toBeDefined();
                expect(response.choices[0].message.content).toBeDefined();

                if (response.choices[0].message.parsed) {
                    expect(typeof response.choices[0].message.parsed).toBe('object');
                } else {
                    expect(() => JSON.parse(response.choices[0].message.content)).not.toThrow();
                }

                if (response.usage) {
                    expect(response.usage.total_tokens).toBeGreaterThan(0);
                }
            } finally {
                try {
                    await client.evals.extract.delete(evalId);
                } catch (error) {
                    // ignore cleanup errors
                }
            }
        }, { timeout: TEST_TIMEOUT });

        test('test_projects_split_uses_split_project_config', async () => {
            const splitEvalName = `test_split_project_${generateId()}`;
            const splitDocumentPath = getFidelityFormPath();

            const splitEval = await client.evals.split.create({
                name: splitEvalName,
                split_config: [
                    { name: 'Document', description: 'The full document' },
                ],
            });
            const evalId = splitEval.id;

            try {
                const response: any = await client.evals.split.process({
                    eval_id: evalId,
                    document: splitDocumentPath,
                });

                expect(response).toBeDefined();
                expect(response.choices).toBeDefined();
                expect(response.choices.length).toBeGreaterThan(0);
                expect(response.choices[0].message.content).toBeDefined();

                const parsed = JSON.parse(response.choices[0].message.content);
                expect(parsed).toBeDefined();
                expect(parsed.splits).toBeDefined();
            } finally {
                try {
                    await client.evals.split.delete(evalId);
                } catch (error) {
                    // ignore cleanup errors
                }
            }
        }, { timeout: TEST_TIMEOUT });

        test('test_projects_extract_functionality', async () => {
            const evaluationName = `test_extract_${generateId()}`;

            const evaluation = await client.evals.extract.create({
                name: evaluationName,
                json_schema: bookingConfirmationJsonSchema,
            });
            const evalId = evaluation.id;

            await client.evals.extract.publish(evalId);

            try {
                const response: any = await client.evals.extract.process({
                    eval_id: evalId,
                    document: bookingConfirmationFilePath2,
                });

                expect(response.id).toBeDefined();
                expect(response.extraction_id).toBeDefined();
                expect(response.choices.length).toBeGreaterThan(0);
                expect(response.choices[0].message.content).toBeDefined();

                if (response.choices[0].message.parsed) {
                    expect(typeof response.choices[0].message.parsed).toBe('object');
                } else {
                    const parsed = JSON.parse(response.choices[0].message.content);
                    expect(typeof parsed).toBe('object');
                }
            } finally {
                try {
                    await client.evals.extract.delete(evalId);
                } catch (error) {
                    // Ignore cleanup errors
                }
            }
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Basic Evaluation CRUD Operations', () => {
        test('test_evaluation_crud_basic', async () => {
            const evaluationName = `test_eval_basic_${generateId()}`;

            const evaluation = await client.evals.extract.create({
                name: evaluationName,
                json_schema: bookingConfirmationJsonSchema,
            });

            expect(evaluation.name).toBe(evaluationName);
            expect(evaluation.draft_config.json_schema).toEqual(bookingConfirmationJsonSchema);

            const evalId = evaluation.id;

            try {
                const retrievedEvaluation = await client.evals.extract.get(evalId);
                expect(retrievedEvaluation.id).toBe(evalId);
                expect(retrievedEvaluation.name).toBe(evaluationName);

                const publishedEvaluation = await client.evals.extract.publish(evalId);
                expect(publishedEvaluation.id).toBe(evalId);

                const evaluations = await client.evals.extract.list();
                expect(evaluations.some(e => e.id === evalId)).toBe(true);
            } finally {
                try {
                    await client.evals.extract.delete(evalId);
                } catch (error) {
                    // Ignore cleanup errors
                }
            }
        }, { timeout: TEST_TIMEOUT });
    });
});
