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
type Modality = "text" | "image" | "native";
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
                document: document,
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
                document: document,
                model: model,
            });
        }
    } else if (clientType === "async") {
        // For TypeScript/Node.js, we don't have separate sync/async clients like Python
        // So we'll just use the same client for both cases
        if (responseMode === "stream") {
            const streamIterator = await client.documents.extract_stream({
                json_schema: jsonSchema,
                document: document,
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
                document: document,
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

    describe('Extract Response Structure Validation', () => {
        test('test_extract_response_structure', async () => {
            const response = await client.documents.extract({
                json_schema: bookingConfirmationJsonSchema,
                document: bookingConfirmationFilePath1,
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
                document: bookingConfirmationFilePath1,
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
                    document: bookingConfirmationFilePath1,
                    model: model,
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
                    document: bookingConfirmationFilePath1,
                    model: "gpt-4.1-nano",
                    reasoning_effort: effort,
                });

                validateExtractionResponse(response);
                expect(response.choices[0].message.content).toBeDefined();
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Extract with Additional Messages', () => {
        test('test_extract_with_text_additional_message', async () => {
            // Test extraction with a text-only additional message providing context
            const additionalMessages = [
                {
                    role: "user" as const,
                    content: "Important context: Please extract all booking details carefully, paying attention to dates and confirmation numbers."
                }
            ];

            const response = await client.documents.extract({
                json_schema: bookingConfirmationJsonSchema,
                document: bookingConfirmationFilePath1,
                model: "gpt-4.1-nano",
                additional_messages: additionalMessages,
            });

            validateExtractionResponse(response);
            expect(response.choices[0].message.content).toBeDefined();
        }, { timeout: TEST_TIMEOUT });

        test('test_extract_with_multipart_additional_message', async () => {
            // Test extraction with additional messages containing both text and image URL
            // Similar to the Python example with Carmoola logo
            const additionalMessages = [
                {
                    role: "user" as const,
                    content: "Context for extraction: This is a booking confirmation document. Please extract all relevant fields."
                },
                {
                    role: "user" as const,
                    content: [
                        {
                            type: "text" as const,
                            text: "Reference image for brand identification:"
                        },
                        {
                            type: "image_url" as const,
                            image_url: {
                                url: "https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png",
                                detail: "auto" as const
                            }
                        }
                    ]
                }
            ];

            const response = await client.documents.extract({
                json_schema: bookingConfirmationJsonSchema,
                document: bookingConfirmationFilePath1,
                model: "gpt-4.1-nano",
                additional_messages: additionalMessages,
            });

            validateExtractionResponse(response);
            expect(response.choices[0].message.content).toBeDefined();
        }, { timeout: TEST_TIMEOUT });

        test('test_extract_with_system_additional_message', async () => {
            // Test extraction with a system message to set behavior
            const additionalMessages = [
                {
                    role: "system" as const,
                    content: "You are a precise document extraction assistant. Extract information exactly as it appears in the document without making assumptions."
                },
                {
                    role: "user" as const,
                    content: "Please be thorough in extracting all booking confirmation details."
                }
            ];

            const response = await client.documents.extract({
                json_schema: bookingConfirmationJsonSchema,
                document: bookingConfirmationFilePath1,
                model: "gpt-4.1-nano",
                additional_messages: additionalMessages,
            });

            validateExtractionResponse(response);
            expect(response.choices[0].message.content).toBeDefined();
        }, { timeout: TEST_TIMEOUT });

        test('test_extract_with_multiple_additional_messages', async () => {
            // Test extraction with multiple additional messages of different types
            const additionalMessages = [
                {
                    role: "developer" as const,
                    content: "Extract data with high precision. When uncertain, prefer leaving fields empty rather than guessing."
                },
                {
                    role: "user" as const,
                    content: "This booking confirmation is from a travel agency. Look for: confirmation number, dates, guest names, and total amount."
                },
                {
                    role: "user" as const,
                    content: "If any field is unclear or ambiguous, please indicate that in your extraction."
                }
            ];

            const response = await client.documents.extract({
                json_schema: bookingConfirmationJsonSchema,
                document: bookingConfirmationFilePath1,
                model: "gpt-4.1-nano",
                additional_messages: additionalMessages,
            });

            validateExtractionResponse(response);
            expect(response.choices[0].message.content).toBeDefined();
        }, { timeout: TEST_TIMEOUT });
    });
});
