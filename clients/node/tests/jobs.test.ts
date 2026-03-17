/**
 * Integration tests for the Jobs API across all supported endpoints.
 *
 * Tests exercise the full job lifecycle: create → poll → verify result.
 *
 * Run:
 *   cd open-source/sdk/clients/node
 *   NODE_ENV=local bun test tests/jobs.test.ts
 */

import { describe, test, beforeAll, expect } from 'bun:test';
import { Retab } from '../src/index.js';
import { getEnvConfig } from './fixtures';
import type { Job } from '../src/api/jobs/client.js';

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const TEST_TIMEOUT = 180_000; // 3 minutes per test
const JOB_TIMEOUT_MS = 120_000; // 2 minutes for job polling
const POLL_INTERVAL_MS = 2_000;
const MODEL = 'retab-micro';

const INLINE_TEXT_DOCUMENT = {
    filename: 'test_invoice.txt',
    url:
        'data:text/plain;base64,' +
        Buffer.from(
            'Invoice #12345\nDate: 2025-01-15\nAmount: $99.99\nCustomer: Acme Corp\nDescription: Consulting services',
        ).toString('base64'),
};

const SIMPLE_EXTRACT_SCHEMA = {
    type: 'object',
    properties: {
        invoice_number: { type: 'string', description: 'The invoice number' },
        amount: { type: 'string', description: 'The total amount' },
        customer: { type: 'string', description: 'The customer name' },
    },
    required: ['invoice_number', 'amount', 'customer'],
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

async function waitAndAssertCompleted(client: Retab, jobId: string): Promise<Job> {
    const job = await client.jobs.waitForCompletion(jobId, {
        poll_interval_ms: POLL_INTERVAL_MS,
        timeout_ms: JOB_TIMEOUT_MS,
        include_response: true,
    });
    expect(job.status).toBe('completed');
    expect(job.response).toBeDefined();
    expect(job.response!.status_code).toBe(200);
    if (job.status !== 'completed') {
        console.error('Job failed:', JSON.stringify(job.error, null, 2));
    }
    return job;
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('Jobs API', () => {
    let client: Retab;

    beforeAll(() => {
        const envConfig = getEnvConfig();
        client = new Retab({
            apiKey: envConfig.retabApiKey,
            baseUrl: envConfig.retabApiBaseUrl,
        });
    });

    // -----------------------------------------------------------------------
    // Per-endpoint completion tests
    // -----------------------------------------------------------------------

    describe('endpoint completion', () => {
        test(
            'extract — job completes and returns choices',
            async () => {
                const job = await client.jobs.create({
                    endpoint: '/v1/documents/extract',
                    request: {
                        document: INLINE_TEXT_DOCUMENT,
                        json_schema: SIMPLE_EXTRACT_SCHEMA,
                        model: MODEL,
                    },
                });
                expect(['queued', 'validating']).toContain(job.status);

                const completed = await waitAndAssertCompleted(client, job.id);
                expect(completed.response!.body).toHaveProperty('choices');
            },
            { timeout: TEST_TIMEOUT },
        );

        test(
            'parse — job completes and returns text',
            async () => {
                const job = await client.jobs.create({
                    endpoint: '/v1/documents/parse',
                    request: {
                        document: INLINE_TEXT_DOCUMENT,
                        model: MODEL,
                    },
                });
                expect(['queued', 'validating']).toContain(job.status);

                const completed = await waitAndAssertCompleted(client, job.id);
                const body = completed.response!.body;
                const hasContent = 'text' in body || 'pages' in body;
                expect(hasContent).toBe(true);
            },
            { timeout: TEST_TIMEOUT },
        );

        test(
            'split — job completes and returns splits',
            async () => {
                const job = await client.jobs.create({
                    endpoint: '/v1/documents/split',
                    request: {
                        document: INLINE_TEXT_DOCUMENT,
                        subdocuments: [
                            { name: 'invoice_header', description: 'The header section with invoice number and date' },
                            { name: 'invoice_details', description: 'The details section with amount and customer' },
                        ],
                        model: MODEL,
                    },
                });
                expect(['queued', 'validating']).toContain(job.status);

                const completed = await waitAndAssertCompleted(client, job.id);
                expect(completed.response!.body).toHaveProperty('splits');
                expect(Array.isArray(completed.response!.body.splits)).toBe(true);
            },
            { timeout: TEST_TIMEOUT },
        );

        test(
            'classify — job completes and returns result',
            async () => {
                const job = await client.jobs.create({
                    endpoint: '/v1/documents/classify',
                    request: {
                        document: INLINE_TEXT_DOCUMENT,
                        categories: [
                            { name: 'invoice', description: 'A billing invoice document' },
                            { name: 'receipt', description: 'A payment receipt' },
                            { name: 'contract', description: 'A legal contract' },
                        ],
                        model: MODEL,
                    },
                });
                expect(['queued', 'validating']).toContain(job.status);

                const completed = await waitAndAssertCompleted(client, job.id);
                expect(completed.response!.body).toHaveProperty('result');
                expect(completed.response!.body.result).toHaveProperty('classification');
            },
            { timeout: TEST_TIMEOUT },
        );

        test(
            'schema generate — job reaches terminal state',
            async () => {
                const job = await client.jobs.create({
                    endpoint: '/v1/schemas/generate',
                    request: {
                        documents: [INLINE_TEXT_DOCUMENT],
                        model: MODEL,
                    },
                });
                expect(['queued', 'validating']).toContain(job.status);

                const finished = await client.jobs.waitForCompletion(job.id, {
                    poll_interval_ms: POLL_INTERVAL_MS,
                    timeout_ms: JOB_TIMEOUT_MS,
                    include_response: true,
                });
                // Schema generation can fail with lightweight models producing invalid schemas
                expect(['completed', 'failed']).toContain(finished.status);
                if (finished.status === 'completed') {
                    expect(finished.response).toBeDefined();
                    expect(typeof finished.response!.body).toBe('object');
                }
            },
            { timeout: TEST_TIMEOUT },
        );
    });

    // -----------------------------------------------------------------------
    // Lifecycle tests
    // -----------------------------------------------------------------------

    describe('lifecycle', () => {
        test(
            'create returns queued job',
            async () => {
                const job = await client.jobs.create({
                    endpoint: '/v1/documents/parse',
                    request: {
                        document: INLINE_TEXT_DOCUMENT,
                        model: MODEL,
                    },
                });
                expect(job.id).toBeDefined();
                expect(['queued', 'validating']).toContain(job.status);
                expect(job.endpoint).toBe('/v1/documents/parse');
                expect(job.object).toBe('job');

                // Clean up
                await client.jobs.waitForCompletion(job.id, { timeout_ms: JOB_TIMEOUT_MS });
            },
            { timeout: TEST_TIMEOUT },
        );

        test(
            'retrieve without payload omits request and response',
            async () => {
                const job = await client.jobs.create({
                    endpoint: '/v1/documents/parse',
                    request: {
                        document: INLINE_TEXT_DOCUMENT,
                        model: MODEL,
                    },
                });
                await client.jobs.waitForCompletion(job.id, { timeout_ms: JOB_TIMEOUT_MS });

                const retrieved = await client.jobs.retrieve(job.id);
                expect(retrieved.id).toBe(job.id);
                expect(retrieved.status).toBe('completed');
                expect(retrieved.request).toBeFalsy();
                expect(retrieved.response).toBeFalsy();
            },
            { timeout: TEST_TIMEOUT },
        );

        test(
            'retrieveFull includes request and response',
            async () => {
                const job = await client.jobs.create({
                    endpoint: '/v1/documents/parse',
                    request: {
                        document: INLINE_TEXT_DOCUMENT,
                        model: MODEL,
                    },
                });
                await client.jobs.waitForCompletion(job.id, { timeout_ms: JOB_TIMEOUT_MS });

                const full = await client.jobs.retrieveFull(job.id);
                expect(full.id).toBe(job.id);
                expect(full.status).toBe('completed');
                expect(full.request).toBeDefined();
                expect(full.response).toBeDefined();
                expect(full.response!.status_code).toBe(200);
            },
            { timeout: TEST_TIMEOUT },
        );

        test(
            'list with filters returns matching jobs',
            async () => {
                const job = await client.jobs.create({
                    endpoint: '/v1/documents/parse',
                    request: {
                        document: INLINE_TEXT_DOCUMENT,
                        model: MODEL,
                    },
                });
                await client.jobs.waitForCompletion(job.id, { timeout_ms: JOB_TIMEOUT_MS });

                const result = await client.jobs.list({
                    endpoint: '/v1/documents/parse',
                    status: 'completed',
                    limit: 5,
                });
                expect(result.data.length).toBeGreaterThan(0);
                for (const j of result.data) {
                    expect(j.endpoint).toBe('/v1/documents/parse');
                    expect(j.status).toBe('completed');
                }
            },
            { timeout: TEST_TIMEOUT },
        );

        test(
            'metadata roundtrip',
            async () => {
                const metadata = { test_key: 'test_value', source: 'node_sdk_test' };
                const job = await client.jobs.create({
                    endpoint: '/v1/documents/parse',
                    request: {
                        document: INLINE_TEXT_DOCUMENT,
                        model: MODEL,
                    },
                    metadata,
                });
                expect(job.metadata).toEqual(metadata);

                await client.jobs.waitForCompletion(job.id, { timeout_ms: JOB_TIMEOUT_MS });

                const retrieved = await client.jobs.retrieve(job.id);
                expect(retrieved.metadata).toEqual(metadata);
            },
            { timeout: TEST_TIMEOUT },
        );

        test(
            'cancel a queued job',
            async () => {
                const job = await client.jobs.create({
                    endpoint: '/v1/documents/parse',
                    request: {
                        document: INLINE_TEXT_DOCUMENT,
                        model: MODEL,
                    },
                });
                try {
                    const cancelled = await client.jobs.cancel(job.id);
                    expect(cancelled.status).toBe('cancelled');
                } catch (e: any) {
                    // Job may have already completed or moved past cancellable state
                    if (e?.response?.status !== 409 && e?.status !== 409) {
                        throw e;
                    }
                }
            },
            { timeout: TEST_TIMEOUT },
        );

        test(
            'invalid endpoint raises error',
            async () => {
                let threw = false;
                try {
                    await client.jobs.create({
                        endpoint: '/v1/nonexistent/endpoint' as any,
                        request: { document: INLINE_TEXT_DOCUMENT },
                    });
                } catch {
                    threw = true;
                }
                expect(threw).toBe(true);
            },
            { timeout: TEST_TIMEOUT },
        );

        test(
            'invalid request body raises error',
            async () => {
                let threw = false;
                try {
                    await client.jobs.create({
                        endpoint: '/v1/documents/extract',
                        request: {
                            // Missing required 'document' and 'json_schema'
                            model: MODEL,
                        },
                    });
                } catch {
                    threw = true;
                }
                expect(threw).toBe(true);
            },
            { timeout: TEST_TIMEOUT },
        );
    });
});
