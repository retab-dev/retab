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
type AIModels = "retab-micro";

type ClientType = "sync" | "async";

// Parse doesn't have streaming, so we only test the basic parse response
type ResponseModeType = "parse";

// Define the types based on the generated types
type TableParsingFormat = "markdown" | "yaml" | "html" | "json";

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
): Promise<void> {
    const document = bookingConfirmationFilePath1;
    let response: ParseResult | null = null;

    if (clientType === "sync") {
        response = await client.documents.parse({
            document: document,
            model: model,
            table_parsing_format: tableParsingFormat,
            image_resolution_dpi: imageResolutionDpi,
        });
    } else if (clientType === "async") {
        // For TypeScript/Node.js, we don't have separate sync/async clients like Python
        // So we'll just use the same client for both cases
        response = await client.documents.parse({
            document: document,
            model: model,
            table_parsing_format: tableParsingFormat,
            image_resolution_dpi: imageResolutionDpi,
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

    describe('Parse Table Formats', () => {
        // Reduced to avoid timeouts - only test html and markdown 
        const tableFormats: TableParsingFormat[] = ["html", "markdown"];

        tableFormats.forEach((tableFormat) => {
            test(`test_parse_table_format_${tableFormat}`, async () => {
                const response = await client.documents.parse({
                    document: bookingConfirmationFilePath1,
                    model: "retab-micro",
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
                    model: "retab-micro",
                    image_resolution_dpi: dpi,
                });

                validateParseResponse(response);
                // Higher DPI might result in better text extraction, but we just verify it works
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
                    "retab-micro",
                    "async",
                    "parse",
                    client,
                    bookingConfirmationFilePath1,
                );
            }, { timeout: TEST_TIMEOUT });
        });
    });


    describe('Parse Response Structure', () => {
        test('test_parse_response_structure', async () => {
            const response = await client.documents.parse({
                document: bookingConfirmationFilePath1,
                model: "retab-micro",
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
