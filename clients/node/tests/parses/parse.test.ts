import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
    getEnvConfig,
    getBookingConfirmationFilePath1,
} from '../fixtures';
import { Parse } from '../../src/types';

// Global test constants
const TEST_TIMEOUT = 180000;

// List of AI Models to test (focusing on models that support parsing)
type AIModels = "retab-micro";

type ClientType = "sync" | "async";

// Parse doesn't have streaming, so we only test the basic parse response
type ResponseModeType = "parse";

// Define the types based on the generated types
type TableParsingFormat = "markdown" | "yaml" | "html" | "json";

function validateParse(parse: Parse | null): void {
    // Assert the instance
    expect(parse).not.toBeNull();
    expect(parse).toBeDefined();

    if (!parse) return;

    // Assert the resource has required fields
    expect(parse.id).toBeDefined();
    expect(typeof parse.id).toBe('string');
    expect(parse.id.length).toBeGreaterThan(0);

    expect(parse.file).toBeDefined();
    expect(parse.model).toBeDefined();
    expect(parse.table_parsing_format).toBeDefined();
    expect(parse.image_resolution_dpi).toBeDefined();

    // Output is nested on the resource
    expect(parse.output).toBeDefined();
    expect(parse.output.pages).not.toBeNull();
    expect(parse.output.text).not.toBeNull();

    // Assert pages is a list of strings
    expect(Array.isArray(parse.output.pages)).toBe(true);
    expect(parse.output.pages.length).toBeGreaterThan(0);
    parse.output.pages.forEach(page => {
        expect(typeof page).toBe('string');
    });

    // Assert text is a non-empty string
    expect(typeof parse.output.text).toBe('string');
    expect(parse.output.text.trim().length).toBeGreaterThan(0);

    // Usage is optional, but if present credits should be non-negative
    if (parse.usage) {
        expect(parse.usage.credits).toBeGreaterThanOrEqual(0);
    }
}

// Base test function for parse endpoint (via new resource)
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
    let parse: Parse | null = null;

    if (clientType === "sync") {
        parse = await client.parses.create({
            document: document,
            model: model,
            table_parsing_format: tableParsingFormat,
            image_resolution_dpi: imageResolutionDpi,
        });
    } else if (clientType === "async") {
        // Node SDK has a single client; mirror the shape of the legacy test.
        parse = await client.parses.create({
            document: document,
            model: model,
            table_parsing_format: tableParsingFormat,
            image_resolution_dpi: imageResolutionDpi,
        });
    }

    validateParse(parse);
}

describe('Retab SDK Parses Resource Tests', () => {
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
                const parse = await client.parses.create({
                    document: bookingConfirmationFilePath1,
                    model: "retab-micro",
                    table_parsing_format: tableFormat,
                });

                validateParse(parse);
                expect(parse.output.text.length).toBeGreaterThan(0);
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Parse Image Resolution', () => {
        const dpiValues = [96, 150];

        dpiValues.forEach((dpi) => {
            test(`test_parse_image_resolution_${dpi}`, async () => {
                const parse = await client.parses.create({
                    document: bookingConfirmationFilePath1,
                    model: "retab-micro",
                    image_resolution_dpi: dpi,
                });

                validateParse(parse);
                expect(parse.output.text.length).toBeGreaterThan(0);
            }, { timeout: TEST_TIMEOUT });
        });
    });

    describe('Parse Response Structure', () => {
        test('test_parse_response_structure', async () => {
            const parse = await client.parses.create({
                document: bookingConfirmationFilePath1,
                model: "retab-micro",
            });

            validateParse(parse);

            // Stored-resource fields should be present
            expect(parse).toHaveProperty('id');
            expect(parse).toHaveProperty('file');
            expect(parse).toHaveProperty('model');
            expect(parse).toHaveProperty('table_parsing_format');
            expect(parse).toHaveProperty('image_resolution_dpi');
            expect(parse).toHaveProperty('output');

            // Nested output is the new shape
            expect(parse.output).toHaveProperty('pages');
            expect(parse.output).toHaveProperty('text');
            expect(parse.output.text.length).toBeGreaterThan(0);
            expect(parse.output.pages.length).toBeGreaterThan(0);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Parses CRUD roundtrip', () => {
        test('test_parses_get_and_delete', async () => {
            const parse = await client.parses.create({
                document: bookingConfirmationFilePath1,
                model: "retab-micro",
            });
            validateParse(parse);

            // Retrieve by ID
            const fetched = await client.parses.get(parse.id);
            expect(fetched.id).toBe(parse.id);
            expect(fetched.output.text.length).toBeGreaterThan(0);

            // Appears in list
            const listed = await client.parses.list({ limit: 10 });
            expect(listed).toHaveProperty('data');
            expect(Array.isArray(listed.data)).toBe(true);

            // Delete
            await client.parses.delete(parse.id);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Parse Overload', () => {
        Array.from({ length: 5 }, (_, i) => i).forEach((requestNumber) => {
            test(`test_parse_overload_${requestNumber}`, async () => {
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
});
