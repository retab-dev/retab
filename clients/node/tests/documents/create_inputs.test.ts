import { randomBytes } from 'node:crypto';
import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
    getEnvConfig,
    getBookingConfirmationFilePath1,
    getBookingConfirmationJsonSchema,
} from '../fixtures';
import { DocumentMessage } from '../../src/types';

// Simple ID generator to replace nanoid
function generateId(): string {
    return randomBytes(6).toString('hex');
}

// Global test constants
const TEST_TIMEOUT = 60000;

type ClientType = "sync" | "async";

function validateResponse(response: DocumentMessage | null): void {
    // Assert the instance
    expect(response).not.toBeNull();
    expect(response).toBeDefined();

    if (!response) return;

    // Assert the response content is not None
    expect(response.messages).not.toBeNull();
    expect(Array.isArray(response.messages)).toBe(true);

    // Assert each message has "role" and "content"
    response.messages.forEach(message => {
        expect(message).toHaveProperty('role');
        expect(message).toHaveProperty('content');
        expect(['user', 'developer']).toContain(message.role);
    });
}

// Test the create_inputs endpoint
async function baseTestCreateInputs(
    clientType: ClientType,
    client: Retab,
    bookingConfirmationFilePath1: string,
    bookingConfirmationJsonSchema: Record<string, any>,
): Promise<void> {
    const jsonSchema = bookingConfirmationJsonSchema;
    const document = bookingConfirmationFilePath1;
    let response: DocumentMessage | null = null;

    if (clientType === "sync") {
        response = await client.documents.create_inputs({
            json_schema: jsonSchema,
            document: document,
        });
    } else if (clientType === "async") {
        // For TypeScript/Node.js, we don't have separate sync/async clients like Python
        // So we'll just use the same client for both cases
        response = await client.documents.create_inputs({
            json_schema: jsonSchema,
            document: document,
        });
    }

    validateResponse(response);
}

describe('Retab SDK Create Inputs Tests', () => {
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

    describe('Basic Create Inputs Tests', () => {
        test('test_create_inputs_sync', async () => {
            await baseTestCreateInputs(
                "sync",
                client,
                bookingConfirmationFilePath1,
                bookingConfirmationJsonSchema,
            );
        }, { timeout: TEST_TIMEOUT });

        test('test_create_inputs_async', async () => {
            await baseTestCreateInputs(
                "async",
                client,
                bookingConfirmationFilePath1,
                bookingConfirmationJsonSchema,
            );
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Create Inputs with Idempotency', () => {
        test('test_create_inputs_with_idempotency', async () => {
            const idempotencyKey = generateId();

            const responseInitial = await client.documents.create_inputs({
                json_schema: bookingConfirmationJsonSchema,
                document: bookingConfirmationFilePath1,
                idempotency_key: idempotencyKey,
            });

            await new Promise(resolve => setTimeout(resolve, 2000));
            const t0 = Date.now();

            const responseSecond = await client.documents.create_inputs({
                json_schema: bookingConfirmationJsonSchema,
                document: bookingConfirmationFilePath1,
                idempotency_key: idempotencyKey,
            });

            const t1 = Date.now();
            expect(t1 - t0).toBeLessThan(10000); // Should take less than 10 seconds

            expect(responseInitial.messages).toEqual(responseSecond.messages);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Create Inputs Error Scenarios', () => {
        const errorScenarios = [
            "invalid_json_schema",
            "missing_document"
        ];

        errorScenarios.forEach((errorScenario) => {
            test(`test_create_inputs_idempotency_error_${errorScenario}`, async () => {
                const idempotencyKey = generateId();

                let jsonSchema: any = bookingConfirmationJsonSchema;
                let document = bookingConfirmationFilePath1;

                // Set up error scenarios
                if (errorScenario === "invalid_json_schema") {
                    jsonSchema = { "invalid": "schema without required fields" };
                } else if (errorScenario === "missing_document") {
                    document = "/nonexistent/file.pdf";
                }

                let response1: any = null;
                let response2: any = null;
                let raisedException1: Error | null = null;
                let raisedException2: Error | null = null;

                // First request attempt
                try {
                    response1 = await client.documents.create_inputs({
                        json_schema: jsonSchema,
                        document: document,
                        idempotency_key: idempotencyKey,
                    });
                } catch (e) {
                    raisedException1 = e as Error;
                }

                await new Promise(resolve => setTimeout(resolve, 2000));
                const t0 = Date.now();

                // Second request attempt with same idempotency key
                try {
                    response2 = await client.documents.create_inputs({
                        json_schema: jsonSchema,
                        document: document,
                        idempotency_key: idempotencyKey,
                    });
                } catch (e) {
                    raisedException2 = e as Error;
                }

                const t1 = Date.now();
                expect(t1 - t0).toBeLessThan(10000); // Should take less than 10 seconds

                // Verify that both requests behaved consistently (idempotent behavior)
                if (raisedException1 !== null) {
                    expect(raisedException2).not.toBeNull();
                    expect(response1).toBeNull();
                    expect(response2).toBeNull();
                    // Verify that both exceptions are of the same type
                    expect(raisedException1.constructor.name).toBe(raisedException2!.constructor.name);
                } else {
                    // If no exception was raised, both responses should be successful and identical
                    expect(raisedException2).toBeNull();
                    expect(response1).not.toBeNull();
                    expect(response2).not.toBeNull();
                    validateResponse(response1);
                    validateResponse(response2);
                    expect(response1.messages).toEqual(response2.messages);
                }
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Response Structure Validation', () => {
        test('test_create_inputs_response_structure', async () => {
            const response = await client.documents.create_inputs({
                json_schema: bookingConfirmationJsonSchema,
                document: bookingConfirmationFilePath1,
            });

            // Validate basic structure
            validateResponse(response);

            // Additional specific validations
            expect(response).toHaveProperty('messages');
            expect(Array.isArray(response.messages)).toBe(true);
            expect(response.messages.length).toBeGreaterThan(0);

            // Validate message structure
            response.messages.forEach((message, index) => {
                expect(message).toHaveProperty('role');
                expect(message).toHaveProperty('content');
                expect(typeof message.role).toBe('string');
                // Content can be either string or object depending on the message type
                expect(message.content).toBeDefined();
                expect(message.content).not.toBeNull();
                if (typeof message.content === 'string') {
                    expect(message.content.length).toBeGreaterThan(0);
                } else if (typeof message.content === 'object') {
                    expect(Object.keys(message.content).length).toBeGreaterThan(0);
                }
            });
        }, { timeout: TEST_TIMEOUT });
    });
});
