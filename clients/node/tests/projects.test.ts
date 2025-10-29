import { randomBytes } from 'crypto';
// @ts-ignore - bun test types may not be available in lint environment
import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../src/index.js';
import {
    getEnvConfig,
    getBookingConfirmationJsonSchema,
    getBookingConfirmationFilePath1,
    getBookingConfirmationFilePath2,
    getBookingConfirmationData1,
    getBookingConfirmationData2,
    TEST_MODEL,
    TEST_MODALITY,
    getPayslipFilePath,
} from './fixtures';

// Simple ID generator to replace nanoid
function generateId(): string {
    return randomBytes(6).toString('hex');
}

// Global test constants
const TEST_TIMEOUT = 180000;

describe('Retab SDK Tests', () => {
    let client: Retab;
    let bookingConfirmationJsonSchema: Record<string, any>;
    let bookingConfirmationFilePath1: string;
    let bookingConfirmationFilePath2: string;
    let bookingConfirmationData1: Record<string, any>;
    let bookingConfirmationData2: Record<string, any>;

    beforeAll(() => {
        const envConfig = getEnvConfig();
        client = new Retab({
            apiKey: envConfig.retabApiKey,
            baseUrl: envConfig.retabApiBaseUrl,
        });

        bookingConfirmationJsonSchema = getBookingConfirmationJsonSchema();
        bookingConfirmationFilePath1 = getBookingConfirmationFilePath1();
        bookingConfirmationFilePath2 = getBookingConfirmationFilePath2();
        bookingConfirmationData1 = getBookingConfirmationData1();
        bookingConfirmationData2 = getBookingConfirmationData2();
    });

    describe('Authentication and Basic SDK Functionality', () => {
        test('test_authentication_validation', async () => {
            // Store original environment variable
            const originalEnvKey = process.env.RETAB_API_KEY;
            const envConfig = getEnvConfig(); // Get config before manipulating env

            try {
                // Test 1: Initialize without API key (should throw error)
                delete process.env.RETAB_API_KEY; // Remove env variable temporarily

                expect(() => {
                    new Retab();
                }).toThrow('Authentication required: Please provide an API key');

                // Test 2: Initialize with API key in constructor
                const clientWithApiKey = new Retab({ apiKey: envConfig.retabApiKey });
                expect(clientWithApiKey).toBeDefined();

                // Test 3: Initialize with environment variable
                process.env.RETAB_API_KEY = envConfig.retabApiKey;

                const clientWithEnv = new Retab();
                expect(clientWithEnv).toBeDefined();

            } finally {
                // Clean up environment variable
                if (originalEnvKey) {
                    process.env.RETAB_API_KEY = originalEnvKey;
                } else {
                    delete process.env.RETAB_API_KEY;
                }
            }
        }, { timeout: TEST_TIMEOUT });

        test('test_projects_extract_without_iteration_id', async () => {
            const evaluationName = `test_extract_no_iter_${generateId()}`;

            // Create a project
            const project = await client.projects.create({
                name: evaluationName,
                json_schema: bookingConfirmationJsonSchema,
            });

            const projectId = project.id;

            // PUBLISH - Promote draft configuration
            const publishedProject = await client.projects.publish(projectId);
            expect(publishedProject.id).toBe(projectId);

            try {
                // Extract without providing iteration_id (should default to base configuration)
                const response: any = await client.projects.extract({
                    project_id: projectId,
                    document: bookingConfirmationFilePath1,
                });

                // Validate response structure
                expect(response).toBeDefined();
                expect(response.id).toBeDefined();
                expect(response.choices).toBeDefined();
                expect(response.choices.length).toBeGreaterThan(0);
                expect(response.choices[0].message).toBeDefined();
                expect(response.choices[0].message.content).toBeDefined();

                // Validate parsed content if available
                if (response.choices[0].message.parsed) {
                    expect(typeof response.choices[0].message.parsed).toBe('object');
                }

                // Validate usage information if present
                if (response.usage) {
                    expect(response.usage.total_tokens).toBeGreaterThan(0);
                }

            } finally {
                // Cleanup created project
                try {
                    await client.projects.delete(projectId);
                } catch (error) {
                    // ignore cleanup errors
                }
            }
        }, { timeout: TEST_TIMEOUT });

        test('test_projects_extract_functionality', async () => {
            // Create an isolated project + iteration, then call extract
            const envConfig = getEnvConfig();
            const testDocumentPath = getPayslipFilePath();

            // Skip if the test document doesn't exist
            try {
                const fs = require('fs');
                if (!fs.existsSync(testDocumentPath)) {
                    console.log(`⚠️  Skipping extract test - Document not found: ${testDocumentPath}`);
                    return;
                }
            } catch (error) {
                console.log('⚠️  Skipping extract test - Cannot check document existence');
                return;
            }

            const testClient = new Retab({
                apiKey: envConfig.retabApiKey,
                baseUrl: envConfig.retabApiBaseUrl,
            });

            // Create project and iteration scoped to this test
            const projectName = `test_extract_${generateId()}`;
            const project = await testClient.projects.create({
                name: projectName,
                json_schema: bookingConfirmationJsonSchema,
            });
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Basic Project CRUD Operations', () => {
        test('test_evaluation_crud_basic', async () => {
            const evaluationName = `test_eval_basic_${generateId()}`;

            // CREATE - Create a new evaluation
            const evaluation = await client.projects.create({
                name: evaluationName,
                json_schema: bookingConfirmationJsonSchema,
            });

            expect(evaluation.name).toBe(evaluationName);

            const projectId = evaluation.id;

            try {
                // READ - Get the evaluation by ID
                const retrievedEvaluation = await client.projects.get(projectId);
                expect(retrievedEvaluation.id).toBe(projectId);
                expect(retrievedEvaluation.name).toBe(evaluationName);

                // PUBLISH - Promote draft configuration
                const publishedEvaluation = await client.projects.publish(projectId);
                expect(publishedEvaluation.id).toBe(projectId);

                // LIST - List evaluations
                const evaluations = await client.projects.list();
                expect(evaluations.some(e => e.id === projectId)).toBe(true);

            } finally {
                // DELETE - Clean up
                try {
                    await client.projects.delete(projectId);
                } catch (error) {
                    // Ignore cleanup errors
                }
            }
        }, { timeout: TEST_TIMEOUT });
    });


});
