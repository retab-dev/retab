import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
    getEnvConfig,
    getFidelityFormPath,
} from '../fixtures';
import { SplitResponse, SplitResult, Category } from '../../src/generated_types.js';

// Global test constants
const TEST_TIMEOUT = 180000; // 3 minutes for split operations

// Categories for testing
const DEFAULT_CATEGORIES: Category[] = [
    { name: 'form_section', description: 'Form sections with input fields, checkboxes, or signature areas' },
    { name: 'instructions', description: 'Instructions, terms and conditions, or explanatory text' },
    { name: 'header', description: 'Header sections with logos, titles, or document identifiers' },
];

function validateSplitResponse(response: SplitResponse | null, categories: Category[]): void {
    // Assert the instance
    expect(response).not.toBeNull();
    expect(response).toBeDefined();

    if (!response) return;

    // Assert the response has required fields
    expect(response.splits).not.toBeNull();
    expect(response.splits).toBeDefined();

    // Validate splits
    expect(Array.isArray(response.splits)).toBe(true);
    expect(response.splits.length).toBeGreaterThan(0);

    // Get valid category names
    const validCategoryNames = new Set(categories.map(cat => cat.name));

    response.splits.forEach((split: SplitResult) => {
        expect(split.name).toBeDefined();
        expect(split.start_page).toBeDefined();
        expect(split.end_page).toBeDefined();

        // Validate category name is from provided categories
        expect(validCategoryNames.has(split.name)).toBe(true);

        // Validate page numbers
        expect(split.start_page).toBeGreaterThanOrEqual(1);
        expect(split.end_page).toBeGreaterThanOrEqual(split.start_page);
    });
}

function validateSplitsAreOrdered(response: SplitResponse): void {
    if (response.splits.length <= 1) return;

    for (let i = 1; i < response.splits.length; i++) {
        expect(response.splits[i].start_page).toBeGreaterThanOrEqual(
            response.splits[i - 1].start_page
        );
    }
}

function validateSplitsNonOverlapping(response: SplitResponse): void {
    if (response.splits.length <= 1) return;

    // Sort splits by start_page for easier validation
    const sortedSplits = [...response.splits].sort((a, b) => a.start_page - b.start_page);

    for (let i = 1; i < sortedSplits.length; i++) {
        const prevSplit = sortedSplits[i - 1];
        const currSplit = sortedSplits[i];

        // Current split should start after previous split ends
        expect(currSplit.start_page).toBeGreaterThan(prevSplit.end_page);
    }
}

describe('Retab SDK Split Tests', () => {
    let client: Retab;
    let multiPagePdfPath: string;

    beforeAll(() => {
        const envConfig = getEnvConfig();
        client = new Retab({
            apiKey: envConfig.retabApiKey,
            baseUrl: envConfig.retabApiBaseUrl,
        });

        // Use the Fidelity form as a multi-page PDF for testing
        multiPagePdfPath = getFidelityFormPath();
    });

    describe('Split Document', () => {
        test('test_split_document_with_file_path', async () => {
            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'gemini-2.5-flash',
                categories: DEFAULT_CATEGORIES,
            });

            validateSplitResponse(response, DEFAULT_CATEGORIES);
            validateSplitsAreOrdered(response);
        }, { timeout: TEST_TIMEOUT });

        test('test_split_document_with_buffer', async () => {
            const fs = require('fs');
            const fileBuffer = fs.readFileSync(multiPagePdfPath);

            const response = await client.documents.split({
                document: {
                    filename: 'fidelity_original.pdf',
                    url: `data:application/pdf;base64,${fileBuffer.toString('base64')}`,
                },
                model: 'gemini-2.5-flash',
                categories: DEFAULT_CATEGORIES,
            });

            validateSplitResponse(response, DEFAULT_CATEGORIES);
            validateSplitsAreOrdered(response);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Split Response Structure', () => {
        test('test_split_response_structure', async () => {
            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'gemini-2.5-flash',
                categories: DEFAULT_CATEGORIES,
            });

            // Validate basic structure
            validateSplitResponse(response, DEFAULT_CATEGORIES);

            // Additional specific validations
            expect(response).toHaveProperty('splits');

            // Validate splits have proper structure
            response.splits.forEach((split: SplitResult) => {
                expect(split).toHaveProperty('name');
                expect(split).toHaveProperty('start_page');
                expect(split).toHaveProperty('end_page');

                // Validate types
                expect(typeof split.name).toBe('string');
                expect(typeof split.start_page).toBe('number');
                expect(typeof split.end_page).toBe('number');
            });
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Split Page Coverage', () => {
        test('test_split_page_coverage', async () => {
            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'gemini-2.5-flash',
                categories: DEFAULT_CATEGORIES,
            });

            validateSplitResponse(response, DEFAULT_CATEGORIES);

            // Validate that all page numbers are positive
            response.splits.forEach((split: SplitResult) => {
                expect(split.start_page).toBeGreaterThanOrEqual(1);
                expect(split.end_page).toBeGreaterThanOrEqual(1);

                // Calculate number of pages in this split
                const pageCount = split.end_page - split.start_page + 1;
                expect(pageCount).toBeGreaterThanOrEqual(1);
            });
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Split with Minimal Categories', () => {
        test('test_split_with_minimal_categories', async () => {
            const minimalCategories: Category[] = [
                { name: 'document', description: 'Document content' },
            ];

            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'gemini-2.5-flash',
                categories: minimalCategories,
            });

            // With only one category, all pages should be classified as that category
            validateSplitResponse(response, minimalCategories);
            expect(response.splits.length).toBeGreaterThanOrEqual(1);

            response.splits.forEach((split: SplitResult) => {
                expect(split.name).toBe('document');
            });
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Split Discontinuous Categories', () => {
        test('test_split_discontinuous_categories', async () => {
            const categories: Category[] = [
                { name: 'form_fields', description: 'Sections with form input fields' },
                { name: 'text_content', description: 'Sections with explanatory text or instructions' },
            ];

            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'gemini-2.5-flash',
                categories: categories,
            });

            validateSplitResponse(response, categories);

            // Count occurrences of each category
            const categoryCounts: Record<string, number> = {};
            response.splits.forEach((split: SplitResult) => {
                categoryCounts[split.name] = (categoryCounts[split.name] || 0) + 1;
            });

            // It's valid for a category to appear multiple times (discontinuous sections)
            // This test just verifies the response structure is correct
            const totalSplits = Object.values(categoryCounts).reduce((a, b) => a + b, 0);
            expect(totalSplits).toBe(response.splits.length);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Split with Different Category Types', () => {
        test('test_split_with_detailed_categories', async () => {
            const detailedCategories: Category[] = [
                { name: 'personal_info', description: 'Sections containing personal information fields like name, address, SSN' },
                { name: 'financial_info', description: 'Sections containing financial information like account numbers, amounts' },
                { name: 'legal_text', description: 'Legal disclaimers, terms, conditions, and fine print' },
                { name: 'signatures', description: 'Signature lines and certification areas' },
            ];

            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'gemini-2.5-flash',
                categories: detailedCategories,
            });

            validateSplitResponse(response, detailedCategories);

            // Verify all returned categories are from our list
            const validNames = new Set(detailedCategories.map(c => c.name));
            response.splits.forEach((split: SplitResult) => {
                expect(validNames.has(split.name)).toBe(true);
            });
        }, { timeout: TEST_TIMEOUT });
    });
});
