import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
    getEnvConfig,
    getFidelityFormPath,
} from '../fixtures';
import { SplitResponse, SplitResult, Subdocument } from '../../src/generated_types.js';

// Global test constants
const TEST_TIMEOUT = 180000; // 3 minutes for split operations

// Categories for testing
const DEFAULT_SUBDOCUMENTS: Subdocument[] = [
    { name: 'form_section', description: 'Form sections with input fields, checkboxes, or signature areas' },
    { name: 'instructions', description: 'Instructions, terms and conditions, or explanatory text' },
    { name: 'header', description: 'Header sections with logos, titles, or document identifiers' },
];

function validateSplitResponse(response: SplitResponse | null, subdocuments: Subdocument[]): void {
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

    // Get valid subdocument names
    const validSubdocumentNames = new Set(subdocuments.map(subdoc => subdoc.name));

    // Track whether at least one split has pages (not all should be empty)
    let hasNonEmptySplit = false;

    response.splits.forEach((split: SplitResult) => {
        expect(split.name).toBeDefined();
        expect(split.pages).toBeDefined();
        expect(Array.isArray(split.pages)).toBe(true);

        // Validate subdocument name is from provided subdocuments
        expect(validSubdocumentNames.has(split.name)).toBe(true);

        // Validate page numbers (only for non-empty splits)
        if (split.pages.length > 0) {
            hasNonEmptySplit = true;
            split.pages.forEach((page: number) => {
                expect(page).toBeGreaterThanOrEqual(1);
            });
        }
    });

    // At least one split should have pages assigned
    expect(hasNonEmptySplit).toBe(true);
}

function validateSplitsAreOrdered(response: SplitResponse): void {
    // Filter to only splits with pages for ordering validation
    const nonEmptySplits = response.splits.filter(s => s.pages.length > 0);

    if (nonEmptySplits.length <= 1) return;

    for (let i = 1; i < nonEmptySplits.length; i++) {
        const prevFirstPage = nonEmptySplits[i - 1].pages[0];
        const currFirstPage = nonEmptySplits[i].pages[0];
        expect(currFirstPage).toBeGreaterThanOrEqual(prevFirstPage);
    }
}

function validateSplitsNonOverlapping(response: SplitResponse): void {
    // Filter to only splits with pages for overlap validation
    const nonEmptySplits = response.splits.filter(s => s.pages.length > 0);

    if (nonEmptySplits.length <= 1) return;

    // Sort splits by first page for easier validation
    const sortedSplits = [...nonEmptySplits].sort((a, b) => a.pages[0] - b.pages[0]);

    for (let i = 1; i < sortedSplits.length; i++) {
        const prevSplit = sortedSplits[i - 1];
        const currSplit = sortedSplits[i];

        // Current split should start after previous split ends
        const prevLastPage = prevSplit.pages[prevSplit.pages.length - 1];
        const currFirstPage = currSplit.pages[0];
        expect(currFirstPage).toBeGreaterThan(prevLastPage);
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
        test('test_split_subdocuments_with_file_path', async () => {
            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'retab-small',
                subdocuments: DEFAULT_SUBDOCUMENTS,
            });

            validateSplitResponse(response, DEFAULT_SUBDOCUMENTS);
            validateSplitsAreOrdered(response);
        }, { timeout: TEST_TIMEOUT });

        test('test_split_subdocuments_with_buffer', async () => {
            const fs = require('fs');
            const fileBuffer = fs.readFileSync(multiPagePdfPath);

            const response = await client.documents.split({
                document: {
                    filename: 'fidelity_original.pdf',
                    url: `data:application/pdf;base64,${fileBuffer.toString('base64')}`,
                },
                model: 'retab-small',
                subdocuments: DEFAULT_SUBDOCUMENTS,
            });

            validateSplitResponse(response, DEFAULT_SUBDOCUMENTS);
            validateSplitsAreOrdered(response);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Split Subdocument Response Structure', () => {
        test('test_split_subdocuments_response_structure', async () => {
            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'retab-small',
                subdocuments: DEFAULT_SUBDOCUMENTS,
            });

            // Validate basic structure
            validateSplitResponse(response, DEFAULT_SUBDOCUMENTS);

            // Additional specific validations
            expect(response).toHaveProperty('splits');

            // Validate splits have proper structure
            response.splits.forEach((split: SplitResult) => {
                expect(split).toHaveProperty('name');
                expect(split).toHaveProperty('pages');

                // Validate types
                expect(typeof split.name).toBe('string');
                expect(Array.isArray(split.pages)).toBe(true);
                split.pages.forEach((page: number) => {
                    expect(typeof page).toBe('number');
                });
            });
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Split Subdocument Coverage', () => {
        test('test_split_subdocuments_page_coverage', async () => {
            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'retab-small',
                subdocuments: DEFAULT_SUBDOCUMENTS,
            });

            validateSplitResponse(response, DEFAULT_SUBDOCUMENTS);

            // Validate that all page numbers are positive (for non-empty splits)
            response.splits.forEach((split: SplitResult) => {
                split.pages.forEach((page: number) => {
                    expect(page).toBeGreaterThanOrEqual(1);
                });
            });
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Split with Minimal Subdocuments', () => {
        test('test_split_with_minimal_subdocuments', async () => {
            const minimalSubdocuments: Subdocument[] = [
                { name: 'document', description: 'Document content' },
            ];

            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'retab-small',
                subdocuments: minimalSubdocuments,
            });

            // With only one category, all pages should be classified as that category
            validateSplitResponse(response, minimalSubdocuments);
            expect(response.splits.length).toBeGreaterThanOrEqual(1);

            response.splits.forEach((split: SplitResult) => {
                expect(split.name).toBe('document');
            });
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Split Discontinuous Subdocuments', () => {
        test('test_split_discontinuous_subdocuments', async () => {
            const discontinuousSubdocuments: Subdocument[] = [
                { name: 'form_fields', description: 'Sections with form input fields' },
                { name: 'text_content', description: 'Sections with explanatory text or instructions' },
            ];

            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'retab-small',
                subdocuments: discontinuousSubdocuments,
            });

            validateSplitResponse(response, discontinuousSubdocuments);

            // Count occurrences of each subdocument
            const subdocumentCounts: Record<string, number> = {};
            response.splits.forEach((split: SplitResult) => {
                subdocumentCounts[split.name] = (subdocumentCounts[split.name] || 0) + 1;
            });

            // It's valid for a subdocument to appear multiple times (discontinuous sections)
            // This test just verifies the response structure is correct
            const totalSplits = Object.values(subdocumentCounts).reduce((a, b) => a + b, 0);
            expect(totalSplits).toBe(response.splits.length);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Split with Different Subdocument Types', () => {
        test('test_split_subdocuments_with_detailed_subdocuments', async () => {
            const detailedSubdocuments: Subdocument[] = [
                { name: 'personal_info', description: 'Sections containing personal information fields like name, address, SSN' },
                { name: 'financial_info', description: 'Sections containing financial information like account numbers, amounts' },
                { name: 'legal_text', description: 'Legal disclaimers, terms, conditions, and fine print' },
                { name: 'signatures', description: 'Signature lines and certification areas' },
            ];

            const response = await client.documents.split({
                document: multiPagePdfPath,
                model: 'retab-small',
                subdocuments: detailedSubdocuments,
            });

            validateSplitResponse(response, detailedSubdocuments);

            // Verify all returned subdocuments are from our list
            const validNames = new Set(detailedSubdocuments.map(subdoc => subdoc.name));
            response.splits.forEach((split: SplitResult) => {
                expect(validNames.has(split.name)).toBe(true);
            });
        }, { timeout: TEST_TIMEOUT });
    });
});
