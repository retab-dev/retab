import { CompositionClient, RequestOptions } from "../../client.js";
import * as z from "zod";

// Job status type
type JobStatus = "validating" | "queued" | "in_progress" | "completed" | "failed" | "cancelled" | "expired";

// Supported endpoints
type SupportedEndpoint =
    | "/v1/documents/extract"
    | "/v1/documents/parse"
    | "/v1/documents/split"
    | "/v1/documents/classify"
    | "/v1/schemas/generate"
    | "/v1/edit/agent/fill"
    | "/v1/edit/templates/fill"
    | "/v1/edit/templates/generate"
    | "/v1/projects/extract";  // Requires "project_id" in request body

// Job response schema
const ZJobResponse = z.object({
    status_code: z.number(),
    body: z.record(z.any()),
});
type JobResponse = z.infer<typeof ZJobResponse>;

// Job error schema
const ZJobError = z.object({
    code: z.string(),
    message: z.string(),
    details: z.record(z.any()).nullable().optional(),
});
type JobError = z.infer<typeof ZJobError>;

// Job schema
const ZJob = z.object({
    id: z.string(),
    object: z.literal("job"),
    status: z.enum(["validating", "queued", "in_progress", "completed", "failed", "cancelled", "expired"]),
    endpoint: z.enum([
        "/v1/documents/extract",
        "/v1/documents/parse",
        "/v1/documents/split",
        "/v1/documents/classify",
        "/v1/schemas/generate",
        "/v1/edit/agent/fill",
        "/v1/edit/templates/fill",
        "/v1/edit/templates/generate",
        "/v1/projects/extract",
    ]),
    request: z.record(z.any()),
    response: ZJobResponse.nullable().optional(),
    error: ZJobError.nullable().optional(),
    created_at: z.number(),
    started_at: z.number().nullable().optional(),
    completed_at: z.number().nullable().optional(),
    expires_at: z.number(),
    organization_id: z.string(),
    metadata: z.record(z.string()).nullable().optional(),
});
type Job = z.infer<typeof ZJob>;

// Job list response schema
const ZJobListResponse = z.object({
    object: z.literal("list"),
    data: z.array(ZJob),
    first_id: z.string().nullable().optional(),
    last_id: z.string().nullable().optional(),
    has_more: z.boolean(),
});
type JobListResponse = z.infer<typeof ZJobListResponse>;

export default class APIJobs extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Create a new asynchronous job.
     *
     * @param endpoint - The API endpoint to call ("/v1/documents/extract" or "/v1/documents/parse")
     * @param request - The full request body for the target endpoint
     * @param metadata - Optional metadata (max 16 pairs; keys ≤64 chars, values ≤512 chars)
     */
    async create({
        endpoint,
        request,
        metadata,
    }: {
        endpoint: SupportedEndpoint;
        request: Record<string, any>;
        metadata?: Record<string, string>;
    }, options?: RequestOptions): Promise<Job> {
        const body: Record<string, any> = {
            endpoint,
            request,
        };
        if (metadata !== undefined) body.metadata = metadata;

        return this._fetchJson(ZJob, {
            url: "/v1/jobs",
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Retrieve a job by ID.
     */
    async retrieve(job_id: string, options?: RequestOptions): Promise<Job> {
        return this._fetchJson(ZJob, {
            url: `/v1/jobs/${job_id}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Cancel a queued or in-progress job.
     */
    async cancel(job_id: string, options?: RequestOptions): Promise<Job> {
        return this._fetchJson(ZJob, {
            url: `/v1/jobs/${job_id}/cancel`,
            method: "POST",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List jobs with pagination and optional status filtering.
     */
    async list({
        after,
        limit = 20,
        status,
    }: {
        after?: string;
        limit?: number;
        status?: JobStatus;
    } = {}, options?: RequestOptions): Promise<JobListResponse> {
        const params: Record<string, any> = {
            after,
            limit,
            status,
        };

        // Remove undefined values
        const cleanParams = Object.fromEntries(
            Object.entries(params).filter(([_, v]) => v !== undefined)
        );

        return this._fetchJson(ZJobListResponse, {
            url: "/v1/jobs",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }
}

export { Job, JobListResponse, JobStatus, SupportedEndpoint, JobResponse, JobError };
