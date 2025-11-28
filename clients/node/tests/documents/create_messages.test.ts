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
            });

            validateCreateMessagesResponse(response);
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
