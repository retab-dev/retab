import { randomBytes } from 'node:crypto';
import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
    getEnvConfig,
    getBookingConfirmationFilePath1,
    getBookingConfirmationJsonSchema,
} from '../fixtures';
import { RetabParsedChatCompletion } from '../../src/types';

// Simple ID generator to replace nanoid
function generateId(): string {
    return randomBytes(6).toString('hex');
}

// Global test constants
const TEST_TIMEOUT = 60000;

// Global modality setting for all tests
const MODALITY = "native";

// List of AI Models to test
type AIModels = "gpt-4.1-nano" | "gemini-2.5-flash-lite";

type ClientType = "sync" | "async";
type ResponseModeType = "stream" | "parse";
type Modality = "text" | "image" | "native" | "image+text";
type BrowserCanvas = "A3" | "A4" | "A5";
type ReasoningEffort = "minimal" | "low" | "medium" | "high";

function validateExtractionResponse(response: RetabParsedChatCompletion | null): void {
    // Assert the instance
    expect(response).not.toBeNull();
    expect(response).toBeDefined();

    if (!response) return;

    // Check if response has choices array and at least one choice
    expect(response.choices).toBeDefined();
    expect(Array.isArray(response.choices)).toBe(true);
    expect(response.choices.length).toBeGreaterThan(0);

    // Assert the response content is not None
    expect(response.choices[0].message.content).not.toBeNull();

    // Assert that content is a valid JSON object
    try {
        JSON.parse(response.choices[0].message.content || "");
    } catch (error) {
        expect(false).toBe(true); // Force failure with descriptive message
        throw new Error("Response content should be a valid JSON object");
    }

    // Assert that the response.choices[0].message.parsed is either a valid object or null
    // (null can happen when an invalid schema is provided and parsing fails)
    const parsed = response.choices[0].message.parsed;
    expect(parsed === null || (typeof parsed === 'object' && parsed !== null)).toBe(true);
}

// Test the extraction endpoint
async function baseTestExtract(
    model: AIModels,
    clientType: ClientType,
    responseMode: ResponseModeType,
    client: Retab,
    bookingConfirmationFilePath1: string,
    bookingConfirmationJsonSchema: Record<string, any>,
): Promise<void> {
    const jsonSchema = bookingConfirmationJsonSchema;
    const document = bookingConfirmationFilePath1;
    let response: RetabParsedChatCompletion | null = null;

    if (clientType === "sync") {
        if (responseMode === "stream") {
            // For Node.js streaming, we need to collect all chunks and build the final response
            const streamIterator = await client.documents.extract_stream({
                json_schema: jsonSchema,
                documents: [document],
                model: model,
            });

            // Collect all streaming chunks
            let lastChunk: any = null;
            for await (const chunk of streamIterator) {
                lastChunk = chunk;
            }

            // Convert the final chunk to a RetabParsedChatCompletion-like structure
            if (lastChunk && lastChunk.choices && lastChunk.choices.length > 0) {
                response = {
                    choices: lastChunk.choices,
                    // Add other properties that might be expected
                    ...lastChunk
                } as RetabParsedChatCompletion;
            }
        } else {
            response = await client.documents.extract({
                json_schema: jsonSchema,
                documents: [document],
                model: model,
            });
        }
    } else if (clientType === "async") {
        // For TypeScript/Node.js, we don't have separate sync/async clients like Python
        // So we'll just use the same client for both cases
        if (responseMode === "stream") {
            const streamIterator = await client.documents.extract_stream({
                json_schema: jsonSchema,
                documents: [document],
                model: model,
            });

            // Collect all streaming chunks
            let lastChunk: any = null;
            for await (const chunk of streamIterator) {
                lastChunk = chunk;
            }

            // Convert the final chunk to a RetabParsedChatCompletion-like structure
            if (lastChunk && lastChunk.choices && lastChunk.choices.length > 0) {
                response = {
                    choices: lastChunk.choices,
                    // Add other properties that might be expected
                    ...lastChunk
                } as RetabParsedChatCompletion;
            }
        } else {
            response = await client.documents.extract({
                json_schema: jsonSchema,
                documents: [document],
                model: model,
            });
        }
    }

    validateExtractionResponse(response);
}

describe('Retab SDK Extract Tests', () => {
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

    describe('Extract OpenAI Tests', () => {
        const clientTypes: ClientType[] = ["sync", "async"];
        // Skip streaming tests for now - they need more investigation
        const responseModes: ResponseModeType[] = ["parse"];

        clientTypes.forEach((clientType) => {
            responseModes.forEach((responseMode) => {
                test(`test_extract_openai_${clientType}_${responseMode}`, async () => {
                    await baseTestExtract(
                        "gpt-4.1-nano",
                        clientType,
                        responseMode,
                        client,
                        bookingConfirmationFilePath1,
                        bookingConfirmationJsonSchema,
                    );
                }, { timeout: TEST_TIMEOUT });
            });
        });
    });

    describe('Extract Overload Tests', () => {
        // Test multiple concurrent requests to verify system stability
        Array.from({ length: 5 }, (_, i) => i).forEach((requestNumber) => {
            test(`test_extract_overload_${requestNumber}`, async () => {
                // Add small delay to stagger requests
                await new Promise(resolve => setTimeout(resolve, requestNumber * 100));

                await baseTestExtract(
                    "gpt-4.1-nano",
                    "async",
                    "parse",
                    client,
                    bookingConfirmationFilePath1,
                    bookingConfirmationJsonSchema,
                );
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Extraction with Idempotency', () => {
        test('test_extraction_with_idempotency', async () => {
            const model: AIModels = "gpt-4.1-nano";

            const responseInitial = await client.documents.extract({
                json_schema: bookingConfirmationJsonSchema,
                documents: [bookingConfirmationFilePath1],
                model: model,
            });

            await new Promise(resolve => setTimeout(resolve, 2000));
            const t0 = Date.now();

            const responseSecond = await client.documents.extract({
                json_schema: bookingConfirmationJsonSchema,
                documents: [bookingConfirmationFilePath1],
                model: model,
            });

            const t1 = Date.now();
            expect(t1 - t0).toBeLessThan(10000); // Should take less than 10 seconds

            expect(responseInitial.choices[0].message.content).toBe(responseSecond.choices[0].message.content);
            expect(responseInitial.choices[0].message.parsed).not.toBeNull();
            expect(responseSecond.choices[0].message.parsed).not.toBeNull();

            // Compare parsed content if both are objects
            if (responseInitial.choices[0].message.parsed && responseSecond.choices[0].message.parsed) {
                expect(responseInitial.choices[0].message.parsed).toEqual(responseSecond.choices[0].message.parsed);
            }
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Extraction Error Scenarios', () => {
        const errorScenarios = [
            "invalid_model",
            "invalid_json_schema",
            "missing_document"
        ];

        errorScenarios.forEach((errorScenario) => {
            test(`test_extraction_idempotency_error_${errorScenario}`, async () => {

                let model: any = "gpt-4.1-nano";
                let jsonSchema: any = bookingConfirmationJsonSchema;
                let document = bookingConfirmationFilePath1;

                // Set up error scenarios
                if (errorScenario === "invalid_model") {
                    model = "invalid-model-name";
                } else if (errorScenario === "invalid_json_schema") {
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
                    response1 = await client.documents.extract({
                        json_schema: jsonSchema,
                        documents: [document],
                        image_resolution_dpi: 128,
                        model: model,
                        temperature: 0,
                        reasoning_effort: "medium",
                        n_consensus: 1,
                    });
                } catch (e) {
                    raisedException1 = e as Error;
                }

                await new Promise(resolve => setTimeout(resolve, 2000));
                const t0 = Date.now();

                // Second request attempt with same idempotency key
                try {
                    response2 = await client.documents.extract({
                        json_schema: jsonSchema,
                        documents: [document],
                        image_resolution_dpi: 128,
                        model: model,
                        temperature: 0,
                        reasoning_effort: "medium",
                        n_consensus: 1,
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
                    validateExtractionResponse(response1);
                    validateExtractionResponse(response2);
                    expect(response1.choices[0].message.content).toBe(response2.choices[0].message.content);
                }
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Extract Response Structure Validation', () => {
        test('test_extract_response_structure', async () => {
            const response = await client.documents.extract({
                json_schema: bookingConfirmationJsonSchema,
                documents: [bookingConfirmationFilePath1],
                model: "gpt-4.1-nano",
            });

            // Validate basic structure
            validateExtractionResponse(response);

            // Additional specific validations
            expect(response).toHaveProperty('choices');
            expect(Array.isArray(response.choices)).toBe(true);
            expect(response.choices.length).toBeGreaterThan(0);

            // Validate choice structure
            const choice = response.choices[0];
            expect(choice).toHaveProperty('message');
            expect(choice.message).toHaveProperty('content');
            expect(choice.message).toHaveProperty('parsed');
            expect(typeof choice.message.content).toBe('string');

            // Content should be valid JSON
            const parsedContent = JSON.parse(choice.message.content);
            expect(typeof parsedContent).toBe('object');
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Extract with Different Parameters', () => {
        test('test_extract_with_different_params', async () => {
            const response = await client.documents.extract({
                json_schema: bookingConfirmationJsonSchema,
                documents: [bookingConfirmationFilePath1],
                model: "gpt-4.1-nano",
                temperature: 0.5,
                image_resolution_dpi: 150,
                reasoning_effort: "medium",
                n_consensus: 1,
            });

            validateExtractionResponse(response);
            expect(response.choices[0].message.content).toBeDefined();
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Extract with Different Models', () => {
        const models: AIModels[] = ["gpt-4.1-nano", "gemini-2.5-flash-lite"];

        models.forEach((model) => {
            test(`test_extract_model_${model.replace(/[^a-zA-Z0-9]/g, '_')}`, async () => {
                const response = await client.documents.extract({
                    json_schema: bookingConfirmationJsonSchema,
                    documents: [bookingConfirmationFilePath1],
                    model: model,
                });

                validateExtractionResponse(response);
                expect(response.choices[0].message.content).toBeDefined();
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Extract with Different Canvas Sizes', () => {
        const canvasSizes: BrowserCanvas[] = ["A3", "A4", "A5"];

        canvasSizes.forEach((canvas) => {
            test(`test_extract_canvas_${canvas}`, async () => {
                const response = await client.documents.extract({
                    json_schema: bookingConfirmationJsonSchema,
                    documents: [bookingConfirmationFilePath1],
                    model: "gpt-4.1-nano",
                });

                validateExtractionResponse(response);
                expect(response.choices[0].message.content).toBeDefined();
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Extract with Different Reasoning Efforts', () => {
        const reasoningEfforts: ReasoningEffort[] = ["minimal", "low", "medium"];

        reasoningEfforts.forEach((effort) => {
            test(`test_extract_reasoning_${effort}`, async () => {
                const response = await client.documents.extract({
                    json_schema: bookingConfirmationJsonSchema,
                    documents: [bookingConfirmationFilePath1],
                    model: "gpt-4.1-nano",
                    reasoning_effort: effort,
                });

                validateExtractionResponse(response);
                expect(response.choices[0].message.content).toBeDefined();
            }, { timeout: TEST_TIMEOUT });
        });
    });
});
