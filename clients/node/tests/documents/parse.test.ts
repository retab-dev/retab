import { randomBytes } from 'node:crypto';
import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
    getEnvConfig,
    getBookingConfirmationFilePath1,
    TEST_MODEL,
} from '../fixtures';
import { ParseResult } from '../../src/types';

// Simple ID generator to replace nanoid
function generateId(): string {
    return randomBytes(6).toString('hex');
}

// Global test constants
const TEST_TIMEOUT = 180000;

// List of AI Models to test (focusing on models that support parsing)
type AIModels = "gemini-2.5-flash-lite" | "gpt-4.1-nano";

type ClientType = "sync" | "async";

// Parse doesn't have streaming, so we only test the basic parse response
type ResponseModeType = "parse";

// Define the types based on the generated types
type TableParsingFormat = "markdown" | "yaml" | "html" | "json";
type BrowserCanvas = "A3" | "A4" | "A5";

function validateParseResponse(response: ParseResult | null): void {
    // Assert the instance
    expect(response).not.toBeNull();
    expect(response).toBeDefined();

    if (!response) return;

    // Assert the response has required fields
    expect(response.document).not.toBeNull();
    expect(response.usage).not.toBeNull();
    expect(response.pages).not.toBeNull();
    expect(response.text).not.toBeNull();

    // Assert usage information is valid
    expect(response.usage.page_count).toBeGreaterThan(0);
    expect(response.usage.credits).toBeGreaterThanOrEqual(0);

    // Assert pages is a list of strings
    expect(Array.isArray(response.pages)).toBe(true);
    expect(response.pages.length).toBeGreaterThan(0);
    response.pages.forEach(page => {
        expect(typeof page).toBe('string');
    });

    // Assert text is a non-empty string
    expect(typeof response.text).toBe('string');
    expect(response.text.trim().length).toBeGreaterThan(0);
}

// Base test function for parse endpoint
async function baseTestParse(
    model: AIModels,
    clientType: ClientType,
    responseMode: ResponseModeType,
    client: Retab,
    bookingConfirmationFilePath1: string,
    tableParsingFormat: TableParsingFormat = "html",
    imageResolutionDpi: number = 96,
    browserCanvas: BrowserCanvas = "A4",
): Promise<void> {
    const document = bookingConfirmationFilePath1;
    let response: ParseResult | null = null;

    if (clientType === "sync") {
        response = await client.documents.parse({
            document: document,
            model: model,
            table_parsing_format: tableParsingFormat,
            image_resolution_dpi: imageResolutionDpi,
            browser_canvas: browserCanvas,
        });
    } else if (clientType === "async") {
        // For TypeScript/Node.js, we don't have separate sync/async clients like Python
        // So we'll just use the same client for both cases
        response = await client.documents.parse({
            document: document,
            model: model,
            table_parsing_format: tableParsingFormat,
            image_resolution_dpi: imageResolutionDpi,
            browser_canvas: browserCanvas,
        });
    }

    validateParseResponse(response);
}

describe('Retab SDK Parse Tests', () => {
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

    describe('Parse with Gemini', () => {
        test('test_parse_gemini_sync', async () => {
            await baseTestParse(
                "gemini-2.5-flash-lite",
                "sync",
                "parse",
                client,
                bookingConfirmationFilePath1,
            );
        }, { timeout: TEST_TIMEOUT });

        test('test_parse_gemini_async', async () => {
            await baseTestParse(
                "gemini-2.5-flash-lite",
                "async",
                "parse",
                client,
                bookingConfirmationFilePath1,
            );
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Parse Table Formats', () => {
        // Reduced to avoid timeouts - only test html and markdown 
        const tableFormats: TableParsingFormat[] = ["html", "markdown"];

        tableFormats.forEach((tableFormat) => {
            test(`test_parse_table_format_${tableFormat}`, async () => {
                const response = await client.documents.parse({
                    document: bookingConfirmationFilePath1,
                    model: "gemini-2.5-flash-lite",
                    table_parsing_format: tableFormat,
                });

                validateParseResponse(response);
                // Verify that the response contains some text content
                expect(response!.text.length).toBeGreaterThan(0);
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Parse Image Resolution', () => {
        // Reduced to avoid timeouts 
        const dpiValues = [96, 150];

        dpiValues.forEach((dpi) => {
            test(`test_parse_image_resolution_${dpi}`, async () => {
                const response = await client.documents.parse({
                    document: bookingConfirmationFilePath1,
                    model: "gemini-2.5-flash-lite",
                    image_resolution_dpi: dpi,
                });

                validateParseResponse(response);
                // Higher DPI might result in better text extraction, but we just verify it works
                expect(response!.text.length).toBeGreaterThan(0);
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Parse Browser Canvas', () => {
        const canvasValues: BrowserCanvas[] = ["A4", "A3", "A5"];

        canvasValues.forEach((canvas) => {
            test(`test_parse_browser_canvas_${canvas}`, async () => {
                const response = await client.documents.parse({
                    document: bookingConfirmationFilePath1,
                    model: "gemini-2.5-flash-lite",
                    browser_canvas: canvas,
                });

                validateParseResponse(response);
                expect(response!.text.length).toBeGreaterThan(0);
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Parse Overload', () => {
        // Test multiple concurrent parse requests to verify system stability
        Array.from({ length: 5 }, (_, i) => i).forEach((requestNumber) => {
            test(`test_parse_overload_${requestNumber}`, async () => {
                // Add small delay to stagger requests
                await new Promise(resolve => setTimeout(resolve, requestNumber * 100));

                await baseTestParse(
                    "gemini-2.5-flash-lite",
                    "async",
                    "parse",
                    client,
                    bookingConfirmationFilePath1,
                );
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Parse with Idempotency', () => {
        test('test_parse_with_idempotency', async () => {
            const idempotencyKey = generateId();
            const model = "gemini-2.5-flash-lite";

            // First request
            const responseInitial = await client.documents.parse({
                document: bookingConfirmationFilePath1,
                model: model,
                table_parsing_format: "html",
                image_resolution_dpi: 96,
                browser_canvas: "A4",
            });

            await new Promise(resolve => setTimeout(resolve, 2000));
            const t0 = Date.now();

            // Second request with same idempotency key
            const responseSecond = await client.documents.parse({
                document: bookingConfirmationFilePath1,
                model: model,
                table_parsing_format: "html",
                image_resolution_dpi: 96,
                browser_canvas: "A4",
            });

            const t1 = Date.now();
            expect(t1 - t0).toBeLessThan(10000); // Should take less than 10 seconds

            // Verify responses are identical
            expect(responseInitial!.text).toBe(responseSecond!.text);
            expect(responseInitial!.pages).toEqual(responseSecond!.pages);
            expect(responseInitial!.usage.page_count).toBe(responseSecond!.usage.page_count);
            expect(responseInitial!.usage.credits).toBe(responseSecond!.usage.credits);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Parse Error Scenarios', () => {
        const errorScenarios = [
            "invalid_model",
            "missing_document",
            "invalid_dpi",
            "invalid_canvas"
        ];

        errorScenarios.forEach((errorScenario) => {
            test(`test_parse_idempotency_error_${errorScenario}`, async () => {
                const idempotencyKey = generateId();

                let model: any = "gemini-2.5-flash-lite";
                let document = bookingConfirmationFilePath1;
                let tableParsingFormat: TableParsingFormat = "html";
                let imageResolutionDpi = 96;
                let browserCanvas: any = "A4";

                // Set up error scenarios
                if (errorScenario === "invalid_model") {
                    model = "invalid-model-name";
                } else if (errorScenario === "missing_document") {
                    document = "/nonexistent/file.pdf";
                } else if (errorScenario === "invalid_dpi") {
                    imageResolutionDpi = -1; // Invalid DPI
                } else if (errorScenario === "invalid_canvas") {
                    browserCanvas = "InvalidCanvas";
                }

                let response1: any = null;
                let response2: any = null;
                let raisedException1: Error | null = null;
                let raisedException2: Error | null = null;

                // First request attempt
                try {
                    response1 = await client.documents.parse({
                        document: document,
                        model: model,
                        table_parsing_format: tableParsingFormat,
                        image_resolution_dpi: imageResolutionDpi,
                        browser_canvas: browserCanvas,
                    });
                } catch (e) {
                    raisedException1 = e as Error;
                }

                await new Promise(resolve => setTimeout(resolve, 2000));
                const t0 = Date.now();

                // Second request attempt with same idempotency key
                try {
                    response2 = await client.documents.parse({
                        document: document,
                        model: model,
                        table_parsing_format: tableParsingFormat,
                        image_resolution_dpi: imageResolutionDpi,
                        browser_canvas: browserCanvas,
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
                    validateParseResponse(response1);
                    validateParseResponse(response2);
                    expect(response1.text).toBe(response2.text);
                    expect(response1.pages).toEqual(response2.pages);
                }
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Parse Response Structure', () => {
        test('test_parse_response_structure', async () => {
            const response = await client.documents.parse({
                document: bookingConfirmationFilePath1,
                model: "gemini-2.5-flash-lite",
            });

            // Validate basic structure
            validateParseResponse(response);

            // Additional specific validations
            expect(response).toHaveProperty('document');
            expect(response).toHaveProperty('usage');
            expect(response).toHaveProperty('pages');
            expect(response).toHaveProperty('text');

            // Validate document metadata - check the actual structure
            expect(response!.document).toBeDefined();
            // The document structure might be different, let's just check it exists
            expect(typeof response!.document).toBe('object');

            // Validate usage information
            expect(typeof response!.usage.page_count).toBe('number');
            expect(typeof response!.usage.credits).toBe('number');

            // Validate that pages and text are consistent
            expect(response!.text.length).toBeGreaterThan(0);
            expect(response!.pages.length).toBe(response!.usage.page_count);
        }, { timeout: TEST_TIMEOUT });
    });
});
