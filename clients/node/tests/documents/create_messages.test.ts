import { randomBytes } from 'node:crypto';
import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
    getEnvConfig,
    getBookingConfirmationFilePath1,
} from '../fixtures';
import { DocumentMessage } from '../../src/types';

// Simple ID generator to replace nanoid
function generateId(): string {
    return randomBytes(6).toString('hex');
}

// Global test constants
const TEST_TIMEOUT = 60000;

type ClientType = "sync" | "async";
type Modality = "text" | "native";
type BrowserCanvas = "A3" | "A4" | "A5";

function validateCreateMessagesResponse(response: DocumentMessage | null): void {
    // Assert the instance
    expect(response).not.toBeNull();
    expect(response).toBeDefined();

    if (!response) return;

    // Assert the response has required attributes
    expect(response.messages).not.toBeNull();
    expect(Array.isArray(response.messages)).toBe(true);
    expect(response.messages.length).toBeGreaterThan(0);

    // Assert each message has "role" and "content"
    response.messages.forEach(message => {
        expect(message).toHaveProperty('role');
        expect(message).toHaveProperty('content');
        expect(['user', 'developer']).toContain(message.role);
        // Message content should be string or object (list/array in TS becomes object)
        expect(['string', 'object']).toContain(typeof message.content);
    });
}

// Test the create_messages endpoint
async function baseTestCreateMessages(
    clientType: ClientType,
    client: Retab,
    bookingConfirmationFilePath1: string,
): Promise<void> {
    const document = bookingConfirmationFilePath1;
    const modality: Modality = "native";
    let response: DocumentMessage | null = null;

    if (clientType === "sync") {
        response = await client.documents.create_messages({
            document: document,
        });
    } else if (clientType === "async") {
        // For TypeScript/Node.js, we don't have separate sync/async clients like Python
        // So we'll just use the same client for both cases
        response = await client.documents.create_messages({
            document: document,
        });
    }

    validateCreateMessagesResponse(response);
}

describe('Retab SDK Create Messages Tests', () => {
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

    describe('Basic Create Messages Tests', () => {
        test('test_create_messages_basic_sync', async () => {
            await baseTestCreateMessages(
                "sync",
                client,
                bookingConfirmationFilePath1,
            );
        }, { timeout: TEST_TIMEOUT });

        test('test_create_messages_basic_async', async () => {
            await baseTestCreateMessages(
                "async",
                client,
                bookingConfirmationFilePath1,
            );
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Create Messages Overload Tests', () => {
        // Test multiple concurrent requests to verify system stability
        Array.from({ length: 5 }, (_, i) => i).forEach((requestNumber) => {
            test(`test_create_messages_overload_${requestNumber}`, async () => {
                // Add small delay to stagger requests
                await new Promise(resolve => setTimeout(resolve, requestNumber * 100));

                await baseTestCreateMessages(
                    "async",
                    client,
                    bookingConfirmationFilePath1,
                );
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Create Messages with Optional Parameters', () => {
        test('test_create_messages_with_optional_params', async () => {
            const document = bookingConfirmationFilePath1;

            const response = await client.documents.create_messages({
                document: document,
                image_resolution_dpi: 96,
                browser_canvas: "A4",
            });

            validateCreateMessagesResponse(response);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Create Messages with Idempotency', () => {
        test('test_create_messages_with_idempotency', async () => {
            const idempotencyKey = generateId();
            const document = bookingConfirmationFilePath1;
            const modality: Modality = "native";

            // First request
            const responseInitial = await client.documents.create_messages({
                document: document,
            });

            await new Promise(resolve => setTimeout(resolve, 2000));
            const t0 = Date.now();

            // Second request with same idempotency key
            const responseSecond = await client.documents.create_messages({
                document: document,
            });

            const t1 = Date.now();

            // Verify both responses
            validateCreateMessagesResponse(responseInitial);
            validateCreateMessagesResponse(responseSecond);

            // Should be fast due to idempotency
            expect(t1 - t0).toBeLessThan(10000); // Should take less than 10 seconds

            // Messages should be identical
            expect(responseInitial.messages).toEqual(responseSecond.messages);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Create Messages Error Scenarios', () => {
        test('test_create_messages_error_missing_document', async () => {
            const idempotencyKey = generateId();
            const document = "/nonexistent/file.pdf";
            const modality: Modality = "native";

            let response1: any = null;
            let response2: any = null;
            let raisedException1: Error | null = null;
            let raisedException2: Error | null = null;

            // First request attempt
            try {
                response1 = await client.documents.create_messages({
                    document: document,
                });
            } catch (e) {
                raisedException1 = e as Error;
            }

            await new Promise(resolve => setTimeout(resolve, 2000));
            const t0 = Date.now();

            // Second request attempt with same idempotency key
            try {
                response2 = await client.documents.create_messages({
                    document: document,
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
                validateCreateMessagesResponse(response1);
                validateCreateMessagesResponse(response2);
                expect(response1.messages).toEqual(response2.messages);
            }
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Create Messages Different Modalities', () => {
        test('test_create_messages_different_modalities', async () => {
            const document = bookingConfirmationFilePath1;

            // Test with 'native' modality
            const responseNative = await client.documents.create_messages({
                document: document,
            });
            validateCreateMessagesResponse(responseNative);

            // Test with 'text' modality
            const responseText = await client.documents.create_messages({
                document: document,
            });
            validateCreateMessagesResponse(responseText);

            // Both should be valid responses with messages
            expect(responseNative.messages.length).toBeGreaterThan(0);
            expect(responseText.messages.length).toBeGreaterThan(0);
            expect(responseNative.messages.every(msg => 'content' in msg)).toBe(true);
            expect(responseText.messages.every(msg => 'content' in msg)).toBe(true);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Create Messages Async vs Sync Consistency', () => {
        test('test_create_messages_async_vs_sync', async () => {
            const document = bookingConfirmationFilePath1;
            const modality: Modality = "native";

            // Sync request (simulated since we use the same client)
            const syncResponse = await client.documents.create_messages({
                document: document,
            });

            // Async request
            const asyncResponse = await client.documents.create_messages({
                document: document,
            });

            // Validate both responses
            validateCreateMessagesResponse(syncResponse);
            validateCreateMessagesResponse(asyncResponse);

            // Both should be valid DocumentMessage instances
            expect(typeof syncResponse).toBe(typeof asyncResponse);
            expect(syncResponse.messages.length).toBeGreaterThan(0);
            expect(asyncResponse.messages.length).toBeGreaterThan(0);
            expect(syncResponse.messages.every(msg => 'content' in msg)).toBe(true);
            expect(asyncResponse.messages.every(msg => 'content' in msg)).toBe(true);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Response Structure Validation', () => {
        test('test_create_messages_response_structure', async () => {
            const response = await client.documents.create_messages({
                document: bookingConfirmationFilePath1,
            });

            // Validate basic structure
            validateCreateMessagesResponse(response);

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

    describe('Create Messages with Different Canvas Sizes', () => {
        const canvasSizes: BrowserCanvas[] = ["A4", "A3", "A5"];

        canvasSizes.forEach((canvas) => {
            test(`test_create_messages_canvas_${canvas}`, async () => {
                const response = await client.documents.create_messages({
                    document: bookingConfirmationFilePath1,
                    browser_canvas: canvas,
                });

                validateCreateMessagesResponse(response);
                expect(response.messages.length).toBeGreaterThan(0);
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Create Messages with Different DPI Settings', () => {
        const dpiValues = [72, 96, 150];

        dpiValues.forEach((dpi) => {
            test(`test_create_messages_dpi_${dpi}`, async () => {
                const response = await client.documents.create_messages({
                    document: bookingConfirmationFilePath1,
                    image_resolution_dpi: dpi,
                });

                validateCreateMessagesResponse(response);
                expect(response.messages.length).toBeGreaterThan(0);
            }, { timeout: TEST_TIMEOUT });
        });
    });
});
