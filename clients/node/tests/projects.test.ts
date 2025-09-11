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

            try {
                const iteration = await testClient.projects.iterations.create(project.id, {
                    model: TEST_MODEL,
                    temperature: 0.0,
                    modality: TEST_MODALITY,
                });

                console.log(`🚀 Testing extract with project: ${project.id}, iteration: ${iteration.id}`);

                // Add a timeout promise to prevent hanging
                const extractPromise = testClient.projects.extract({
                    project_id: project.id,
                    iteration_id: iteration.id,
                    document: testDocumentPath,
                });

                const timeoutPromise = new Promise((_, reject) => {
                    setTimeout(() => reject(new Error(`Extract API call timed out after ${TEST_TIMEOUT / 1000} seconds`)), TEST_TIMEOUT);
                });

                const response = await Promise.race([extractPromise, timeoutPromise]) as any;

                // Validate response structure
                expect(response).toBeDefined();
                expect(response.id).toBeDefined();
                expect(response.choices).toBeDefined();
                expect(response.choices.length).toBeGreaterThan(0);
                expect(response.choices[0].message).toBeDefined();
                expect(response.choices[0].message.content).toBeDefined();
                expect(response.extraction_id).toBeDefined();
                expect(response.object).toBe('chat.completion');

                // Validate parsed content if available
                if (response.choices[0].message.parsed) {
                    expect(typeof response.choices[0].message.parsed).toBe('object');
                }

                // Validate usage information
                if (response.usage) {
                    expect(response.usage.total_tokens).toBeGreaterThan(0);
                    expect(response.usage.completion_tokens).toBeGreaterThan(0);
                    expect(response.usage.prompt_tokens).toBeGreaterThan(0);
                }

                console.log(`✅ Extract test successful - Extraction ID: ${response.extraction_id}`);
            } finally {
                // Cleanup created project
                try {
                    await testClient.projects.delete(project.id);
                } catch (_) {
                    // ignore cleanup errors
                }
            }
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
            expect(evaluation.json_schema).toEqual(bookingConfirmationJsonSchema);
            expect(evaluation.documents).toHaveLength(0);
            expect(evaluation.iterations).toHaveLength(0);

            const projectId = evaluation.id;

            try {
                // READ - Get the evaluation by ID
                const retrievedEvaluation = await client.projects.get(projectId);
                expect(retrievedEvaluation.id).toBe(projectId);
                expect(retrievedEvaluation.name).toBe(evaluationName);

                // LIST - List evaluations
                const evaluations = await client.projects.list();
                expect(evaluations.some(e => e.id === projectId)).toBe(true);

                // UPDATE - Update the evaluation
                const updatedName = `updated_${evaluationName}`;
                const updatedEvaluation = await client.projects.update(projectId, {
                    name: updatedName,
                });
                expect(updatedEvaluation.name).toBe(updatedName);

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

    describe('Project with Documents', () => {
        test('test_evaluation_with_documents', async () => {
            const evaluationName = `test_eval_docs_${generateId()}`;

            // Create an evaluation
            const evaluation = await client.projects.create({
                name: evaluationName,
                json_schema: bookingConfirmationJsonSchema,
            });

            const projectId = evaluation.id;

            try {
                // CREATE - Add a document
                const document = await client.projects.documents.create(projectId, {
                    mime_data: bookingConfirmationFilePath1,
                    annotation: bookingConfirmationData1,
                });

                expect(document.annotation).toEqual(bookingConfirmationData1);
                const documentId = document.id;

                // LIST - List documents in the evaluation
                const documents = await client.projects.documents.list(projectId);
                expect(documents).toHaveLength(1);
                expect(documents[0].id).toBe(documentId);

                // UPDATE - Update the document annotation
                const updatedDocument = await client.projects.documents.update(
                    projectId,
                    documentId,
                    {
                        annotation: bookingConfirmationData2,
                    }
                );
                expect(updatedDocument.annotation).toEqual(bookingConfirmationData2);

                // Verify the evaluation still exists and has the document
                const updatedEval = await client.projects.get(projectId);
                expect(updatedEval.documents).toHaveLength(1);

                // DELETE - Remove the document
                await client.projects.documents.delete(projectId, documentId);

                // Verify document was removed
                const documentsAfter = await client.projects.documents.list(projectId);
                expect(documentsAfter).toHaveLength(0);

            } finally {
                // DELETE - Clean up evaluation
                try {
                    await client.projects.delete(projectId);
                } catch (error) {
                    // Ignore cleanup errors
                }
            }
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Iteration CRUD and Processing', () => {
        test('test_iteration_crud_and_processing', async () => {
            const evaluationName = `test_eval_iter_${generateId()}`;

            // First create an evaluation
            const evaluation = await client.projects.create({
                name: evaluationName,
                json_schema: bookingConfirmationJsonSchema,
            });

            const projectId = evaluation.id;

            try {
                // Add a document to the evaluation
                const document = await client.projects.documents.create(projectId, {
                    mime_data: bookingConfirmationFilePath1,
                    annotation: bookingConfirmationData1,
                });
                expect(document.annotation).toEqual(bookingConfirmationData1);

                // CREATE - Create a new iteration
                const iteration = await client.projects.iterations.create(projectId, {
                    model: TEST_MODEL,
                    temperature: 0.1,
                    modality: TEST_MODALITY,
                });

                expect(iteration.inference_settings.model).toBe(TEST_MODEL);
                expect(iteration.inference_settings.temperature).toBe(0.1);
                expect(Object.keys(iteration.predictions)).toHaveLength(1);
                expect(iteration.predictions[document.id].prediction).toEqual({});

                const iterationId = iteration.id;

                // LIST - List iterations for the evaluation
                const iterations = await client.projects.iterations.list(projectId);
                expect(iterations.some(i => i.id === iterationId)).toBe(true);

                // STATUS - Check document status
                const statusResponse = await (client.projects.iterations as any).status(projectId, iterationId);
                expect(statusResponse.documents.length).toBeGreaterThan(0);

                // All documents should need updates initially
                for (const docStatus of statusResponse.documents) {
                    expect(docStatus.needs_update).toBe(true);
                    expect(docStatus.has_prediction).toBe(false);
                }

                // PROCESS - Process the iteration (run extractions)
                const processedIteration = await (client.projects.iterations as any).process(
                    projectId,
                    iterationId,
                    {
                        only_outdated: true,
                    }
                );

                // After processing, predictions should be populated
                expect(Object.keys(processedIteration.predictions).length).toBeGreaterThan(0);

                // Check status again after processing
                const statusAfter = await (client.projects.iterations as any).status(projectId, iterationId);
                for (const docStatus of statusAfter.documents) {
                    expect(docStatus.has_prediction).toBe(true);
                    expect(docStatus.needs_update).toBe(false); // Should not need updates anymore
                }

                // DELETE - Clean up iteration
                try {
                    await client.projects.iterations.delete(projectId, iterationId);
                } catch (error) {
                    // Ignore cleanup errors
                }

            } finally {
                // DELETE - Clean up evaluation
                try {
                    await client.projects.delete(projectId);
                } catch (error) {
                    // Ignore cleanup errors
                }
            }
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Process Document Method', () => {
        test('test_process_document_method', async () => {
            const evaluationName = `test_eval_process_doc_${generateId()}`;

            // Create an evaluation
            const evaluation = await client.projects.create({
                name: evaluationName,
                json_schema: bookingConfirmationJsonSchema,
            });

            const projectId = evaluation.id;

            try {
                // Add a document to the evaluation
                const document = await client.projects.documents.create(projectId, {
                    mime_data: bookingConfirmationFilePath1,
                    annotation: bookingConfirmationData1,
                });

                // Create an iteration
                const iteration = await client.projects.iterations.create(projectId, {
                    model: TEST_MODEL,
                    temperature: 0.0,
                    modality: TEST_MODALITY,
                });

                const iterationId = iteration.id;
                const documentId = document.id;

                // Check initial status - document should need update
                const initialStatus = await (client.projects.iterations as any).status(projectId, iterationId);
                const docStatus = initialStatus.documents.find((d: any) => d.document_id === documentId);
                expect(docStatus.needs_update).toBe(true);
                expect(docStatus.has_prediction).toBe(false);

                // PROCESS_DOCUMENT - Process a single document (frontend method)
                const completionResponse = await (client.projects.iterations as any).process_document(
                    projectId,
                    iterationId,
                    documentId
                );

                // Validate the response
                expect(completionResponse).toBeDefined();
                // Note: API may return { success: true } for process_document
                if (completionResponse.choices) {
                    expect(completionResponse.choices).toBeDefined();
                    expect(completionResponse.choices.length).toBeGreaterThan(0);
                    expect(completionResponse.choices[0].message.content).toBeDefined();
                }

                // Verify that the parsed content is valid JSON (if choices exist)
                if (completionResponse.choices) {
                    try {
                        const parsedContent = JSON.parse(completionResponse.choices[0].message.content);
                        expect(typeof parsedContent).toBe('object');
                    } catch (error) {
                        throw new Error('Response content should be valid JSON');
                    }
                }

                // Check status after processing - document should be updated
                const finalStatus = await (client.projects.iterations as any).status(projectId, iterationId);
                const docStatusAfter = finalStatus.documents.find((d: any) => d.document_id === documentId);
                expect(docStatusAfter.has_prediction).toBe(true);
                expect(docStatusAfter.needs_update).toBe(false);

            } finally {
                // Clean up
                try {
                    await client.projects.delete(projectId);
                } catch (error) {
                    // Ignore cleanup errors
                }
            }
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Complete Evaluation Workflow', () => {
        test('test_complete_evaluation_workflow', async () => {
            const evaluationName = `test_eval_complete_${generateId()}`;

            // Step 1: Create an evaluation
            const evaluation = await client.projects.create({
                name: evaluationName,
                json_schema: bookingConfirmationJsonSchema,
            });

            const projectId = evaluation.id;

            try {
                // Step 2: Add 2 documents
                const doc1 = await client.projects.documents.create(projectId, {
                    mime_data: bookingConfirmationFilePath1,
                    annotation: bookingConfirmationData1,
                });

                const doc2 = await client.projects.documents.create(projectId, {
                    mime_data: bookingConfirmationFilePath2,
                    annotation: bookingConfirmationData2,
                });

                // Verify we have 2 documents
                const documents = await client.projects.documents.list(projectId);
                expect(documents).toHaveLength(2);

                // Step 3: Create 2 iterations
                const iteration1 = await client.projects.iterations.create(projectId, {
                    model: TEST_MODEL,
                    temperature: 0.0,
                    modality: TEST_MODALITY,
                });

                const iteration2 = await client.projects.iterations.create(projectId, {
                    model: TEST_MODEL,
                    temperature: 0.5,
                    modality: TEST_MODALITY,
                });

                // Verify we have 2 iterations
                const iterations = await client.projects.iterations.list(projectId);
                expect(iterations).toHaveLength(2);

                // Step 4: Process using both process methods
                // 4a: Use process method (bulk processing)
                const processedIter1 = await (client.projects.iterations as any).process(
                    projectId,
                    iteration1.id,
                    {
                        only_outdated: true,
                    }
                );

                // 4b: Use process_document method (individual document processing)
                const completion1 = await (client.projects.iterations as any).process_document(
                    projectId,
                    iteration2.id,
                    doc1.id
                );
                const completion2 = await (client.projects.iterations as any).process_document(
                    projectId,
                    iteration2.id,
                    doc2.id
                );

                // Verify both iterations have predictions
                expect(Object.keys(processedIter1.predictions).length).toBeGreaterThan(0);

                // Verify process_document responses
                expect(completion1).toBeDefined();
                expect(completion2).toBeDefined();
                // Note: API may return { success: true } for process_document
                if (completion1.choices) {
                    expect(completion1.choices[0].message.content).toBeDefined();
                }
                if (completion2.choices) {
                    expect(completion2.choices[0].message.content).toBeDefined();
                }

                // Step 5: Delete one iteration
                await client.projects.iterations.delete(projectId, iteration2.id);

                // Verify we now have only 1 iteration
                const iterationsAfterDelete = await client.projects.iterations.list(projectId);
                expect(iterationsAfterDelete).toHaveLength(1);
                expect(iterationsAfterDelete[0].id).toBe(iteration1.id);

                // Step 6: Delete one document (should affect remaining iteration)
                await client.projects.documents.delete(projectId, doc2.id);

                // Verify we now have only 1 document
                const documentsAfterDelete = await client.projects.documents.list(projectId);
                expect(documentsAfterDelete).toHaveLength(1);
                expect(documentsAfterDelete[0].id).toBe(doc1.id);

                // Step 7: Check status of remaining iteration (should reflect document deletion)
                const finalStatus = await (client.projects.iterations as any).status(projectId, iteration1.id);
                expect(finalStatus.documents).toHaveLength(1); // Only one document should remain

            } finally {
                // Cleanup - Delete evaluation (should cascade to remaining documents and iterations)
                try {
                    await client.projects.delete(projectId);
                } catch (error) {
                    // Ignore cleanup errors
                }
            }
        }, { timeout: TEST_TIMEOUT });
    });

    describe('Iteration Selective Processing', () => {
        test('test_iteration_selective_processing', async () => {
            const evaluationName = `test_eval_selective_${generateId()}`;

            // Create an evaluation
            const evaluation = await client.projects.create({
                name: evaluationName,
                json_schema: bookingConfirmationJsonSchema,
            });

            const projectId = evaluation.id;

            try {
                // Add multiple documents
                const doc1 = await client.projects.documents.create(projectId, {
                    mime_data: bookingConfirmationFilePath1,
                    annotation: bookingConfirmationData1,
                });

                const doc2 = await client.projects.documents.create(projectId, {
                    mime_data: bookingConfirmationFilePath2,
                    annotation: bookingConfirmationData2,
                });

                // Create an iteration
                const iteration = await client.projects.iterations.create(projectId, {
                    model: TEST_MODEL,
                    temperature: 0.0,
                    modality: TEST_MODALITY,
                });

                const iterationId = iteration.id;

                // Process only the first document using process method
                await (client.projects.iterations as any).process(
                    projectId,
                    iterationId,
                    {
                        document_ids: [doc1.id],
                        only_outdated: false,
                    }
                );

                // Check that only one document was processed via process method
                const statusResponse = await (client.projects.iterations as any).status(projectId, iterationId);
                const doc1Status = statusResponse.documents.find((d: any) => d.document_id === doc1.id);
                const doc2Status = statusResponse.documents.find((d: any) => d.document_id === doc2.id);

                expect(doc1Status.has_prediction).toBe(true);
                expect(doc2Status.has_prediction).toBe(false);
                expect(doc1Status.needs_update).toBe(false);
                expect(doc2Status.needs_update).toBe(true);

                // Now process the second document using process_document method
                const completionResponse = await (client.projects.iterations as any).process_document(
                    projectId,
                    iterationId,
                    doc2.id
                );

                // Verify the response
                expect(completionResponse).toBeDefined();
                // Note: API may return { success: true } for process_document
                if (completionResponse.choices) {
                    expect(completionResponse.choices[0].message.content).toBeDefined();
                }

                // Check that both documents are now processed
                const finalStatus = await (client.projects.iterations as any).status(projectId, iterationId);
                const doc1Final = finalStatus.documents.find((d: any) => d.document_id === doc1.id);
                const doc2Final = finalStatus.documents.find((d: any) => d.document_id === doc2.id);

                expect(doc1Final.has_prediction).toBe(true);
                expect(doc2Final.has_prediction).toBe(true);
                expect(doc1Final.needs_update).toBe(false);
                expect(doc2Final.needs_update).toBe(false);

                // DELETE - Clean up iteration
                try {
                    await client.projects.iterations.delete(projectId, iterationId);
                } catch (error) {
                    // Ignore cleanup errors
                }

            } finally {
                // DELETE - Clean up evaluation
                try {
                    await client.projects.delete(projectId);
                } catch (error) {
                    // Ignore cleanup errors
                }
            }
        }, { timeout: TEST_TIMEOUT });
    });
});
