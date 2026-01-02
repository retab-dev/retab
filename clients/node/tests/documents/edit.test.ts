import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../../src/index.js';
import {
    getEnvConfig,
    getFidelityFormPath,
    getFidelityInstructions,
    getFidelityInstructionsJson,
} from '../fixtures';
import { EditResponse, FormField, OCRResult, MIMEData } from '../../src/types';

// Global test constants
const TEST_TIMEOUT = 300000; // 5 minutes for edit operations (OCR + LLM)

function validateEditResponse(response: EditResponse | null): void {
    // Assert the instance
    expect(response).not.toBeNull();
    expect(response).toBeDefined();

    if (!response) return;

    expect(response.form_data).not.toBeNull();
    expect(response.form_data).toBeDefined();

    // Validate annotated_pdf is MIMEData with data URI
    const annotatedPdf = response.filled_document as MIMEData;
    expect(annotatedPdf.url).toBeDefined();
    expect(annotatedPdf.url.startsWith('data:')).toBe(true);

    // Validate form_data
    expect(Array.isArray(response.form_data)).toBe(true);
    expect(response.form_data.length).toBeGreaterThan(0);

    response.form_data.forEach((field: FormField) => {
        expect(field.bbox).toBeDefined();
        expect(field.description).toBeDefined();
        expect(field.type).toBeDefined();

        // Validate bbox is normalized (0-1)
        expect(field.bbox.left).toBeGreaterThanOrEqual(0);
        expect(field.bbox.left).toBeLessThanOrEqual(1);
        expect(field.bbox.top).toBeGreaterThanOrEqual(0);
        expect(field.bbox.top).toBeLessThanOrEqual(1);
        expect(field.bbox.width).toBeGreaterThanOrEqual(0);
        expect(field.bbox.width).toBeLessThanOrEqual(1);
        expect(field.bbox.height).toBeGreaterThanOrEqual(0);
        expect(field.bbox.height).toBeLessThanOrEqual(1);
        expect(field.bbox.page).toBeGreaterThanOrEqual(1);
    });

    // Validate filled_document if present
    if (response.filled_document) {
        const filledPdf = response.filled_document as MIMEData;
        expect(filledPdf.filename).toBeDefined();
        expect(filledPdf.url).toBeDefined();
        expect(filledPdf.url.startsWith('data:application/pdf;base64,')).toBe(true);
    }
}

function validateFidelityFormFields(response: EditResponse, instructions: Record<string, any>): void {
    const filledFields = response.form_data.filter((field: FormField) => field.value);

    // Check that some expected values from the instructions are present
    const accountOwners: string[] = instructions.account_owners || [];

    for (const owner of accountOwners) {
        const ownerParts = owner.toUpperCase().split(' ');
        const foundOwner = filledFields.some((field: FormField) => {
            const value = (field.value || '').toUpperCase();
            return ownerParts.some(part => value.includes(part));
        });
        expect(foundOwner).toBe(true);
    }

    // Check for at least one account number
    const accountNumbers: string[] = instructions.fidelity_accounts || [];
    if (accountNumbers.length > 0) {
        const foundAccount = filledFields.some((field: FormField) => {
            const value = field.value || '';
            return accountNumbers.some(acc => value.includes(acc));
        });
        expect(foundAccount).toBe(true);
    }
}

describe('Retab SDK Edit Tests', () => {
    let client: Retab;
    let fidelityFormPath: string;
    let fidelityInstructions: string;
    let fidelityInstructionsJson: Record<string, any>;

    beforeAll(() => {
        const envConfig = getEnvConfig();
        client = new Retab({
            apiKey: envConfig.retabApiKey,
            baseUrl: envConfig.retabApiBaseUrl,
        });

        fidelityFormPath = getFidelityFormPath();
        fidelityInstructions = getFidelityInstructions();
        fidelityInstructionsJson = getFidelityInstructionsJson();
    });

    describe('Edit Fidelity Form', () => {
        test('test_edit_fidelity_form', async () => {
            const response = await client.documents.edit({
                document: {
                    filename: 'fidelity_original.pdf',
                    url: `data:application/pdf;base64,${Buffer.from(
                        require('fs').readFileSync(fidelityFormPath)
                    ).toString('base64')}`,
                },
                instructions: fidelityInstructions,
                model: 'gemini-2.5-flash',
            });

            validateEditResponse(response);
            validateFidelityFormFields(response, fidelityInstructionsJson);
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Edit Response Structure', () => {
        test('test_edit_response_structure', async () => {
            const response = await client.documents.edit({
                document: {
                    filename: 'fidelity_original.pdf',
                    url: `data:application/pdf;base64,${Buffer.from(
                        require('fs').readFileSync(fidelityFormPath)
                    ).toString('base64')}`,
                },
                instructions: fidelityInstructions,
                model: 'gemini-2.5-flash',
            });

            // Validate basic structure
            validateEditResponse(response);

            // Additional specific validations
            expect(response).toHaveProperty('form_data');
            expect(response).toHaveProperty('filled_document');


            // Validate form_data has fields with proper structure
            response.form_data.forEach((field: FormField) => {
                expect(field).toHaveProperty('bbox');
                expect(field).toHaveProperty('description');
                expect(field).toHaveProperty('type');
                expect(field).toHaveProperty('value');
            });
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Edit Filled PDF Validation', () => {
        test('test_edit_filled_document_is_valid', async () => {
            const response = await client.documents.edit({
                document: {
                    filename: 'fidelity_original.pdf',
                    url: `data:application/pdf;base64,${Buffer.from(
                        require('fs').readFileSync(fidelityFormPath)
                    ).toString('base64')}`,
                },
                instructions: fidelityInstructions,
                model: 'gemini-2.5-flash',
            });

            validateEditResponse(response);

            // Extract and validate PDF content
            expect(response.filled_document).not.toBeNull();
            expect(response.filled_document).toBeDefined();

            if (response.filled_document) {
                // Extract base64 content from data URI
                const base64Content = response.filled_document.url.split(',')[1];
                const pdfBuffer = Buffer.from(base64Content, 'base64');

                // Check PDF magic bytes (PDF files start with %PDF-)
                const pdfHeader = pdfBuffer.slice(0, 5).toString('ascii');
                expect(pdfHeader).toBe('%PDF-');
                expect(pdfBuffer.length).toBeGreaterThan(1000);
            }
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Edit Form Data Has Filled Values', () => {
        test('test_edit_form_data_has_filled_values', async () => {
            const response = await client.documents.edit({
                document: {
                    filename: 'fidelity_original.pdf',
                    url: `data:application/pdf;base64,${Buffer.from(
                        require('fs').readFileSync(fidelityFormPath)
                    ).toString('base64')}`,
                },
                instructions: fidelityInstructions,
                model: 'gemini-2.5-flash',
            });

            validateEditResponse(response);

            // Count fields with values
            const filledFields = response.form_data.filter((field: FormField) => field.value);

            expect(filledFields.length).toBeGreaterThan(0);

            // The Fidelity form has multiple fields, expect a reasonable number to be filled
            const totalFields = response.form_data.length;
            const fillRatio = totalFields > 0 ? filledFields.length / totalFields : 0;

            // At least 20% of fields should be filled with the provided instructions
            expect(fillRatio).toBeGreaterThanOrEqual(0.2);
        }, { timeout: TEST_TIMEOUT });
    });
});
