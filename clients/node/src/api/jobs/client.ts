import { CompositionClient, RequestOptions } from "../../client.js";
import * as z from "zod";

// Job status type
type JobStatus = "validating" | "queued" | "in_progress" | "completed" | "failed" | "cancelled" | "expired";
type JobListSource = "api" | "project" | "workflow";
type JobListOrder = "asc" | "desc";

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
    request: z.record(z.any()).nullable().optional(),
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

type JobRetrieveOptions = RequestOptions & {
    include_request?: boolean;
    include_response?: boolean;
};

type JobWaitForCompletionOptions = RequestOptions & {
    poll_interval_ms?: number;
    timeout_ms?: number;
    include_request?: boolean;
    include_response?: boolean;
};

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
    async retrieve(job_id: string, options?: JobRetrieveOptions): Promise<Job> {
        const include_request = options?.include_request ?? false;
        const include_response = options?.include_response ?? false;
        const params: Record<string, any> = {
            ...(options?.params || {}),
            include_request,
            include_response,
        };

        return this._fetchJson(ZJob, {
            url: `/v1/jobs/${job_id}`,
            method: "GET",
            params,
            headers: options?.headers,
        });
    }

    /**
     * Retrieve a job by ID including full request and response payloads.
     */
    async retrieveFull(job_id: string, options?: RequestOptions): Promise<Job> {
        return this.retrieve(job_id, {
            ...options,
            include_request: true,
            include_response: true,
        });
    }

    /**
     * Poll a job until terminal status, then fetch final payload once.
     */
    async waitForCompletion(
        job_id: string,
        options: JobWaitForCompletionOptions = {},
    ): Promise<Job> {
        const poll_interval_ms = options.poll_interval_ms ?? 2000;
        const timeout_ms = options.timeout_ms ?? 600000;
        const include_request = options.include_request ?? false;
        const include_response = options.include_response ?? true;
        if (poll_interval_ms <= 0) throw new Error("poll_interval_ms must be > 0");
        if (timeout_ms <= 0) throw new Error("timeout_ms must be > 0");

        const terminal_statuses = new Set<JobStatus>(["completed", "failed", "cancelled", "expired"]);
        const started_at_ms = Date.now();
        const deadline_ms = started_at_ms + timeout_ms;
        while (true) {
            const job = await this.retrieve(job_id, {
                ...options,
                include_request: false,
                include_response: false,
            });
            if (terminal_statuses.has(job.status)) {
                if (include_request || include_response) {
                    return this.retrieve(job_id, {
                        ...options,
                        include_request,
                        include_response,
                    });
                }
                return job;
            }

            const now_ms = Date.now();
            if (now_ms >= deadline_ms) {
                throw new Error(`Timed out waiting for job ${job_id} completion after ${timeout_ms}ms`);
            }

            const sleep_for_ms = Math.min(poll_interval_ms, Math.max(deadline_ms - now_ms, 0));
            await new Promise((resolve) => setTimeout(resolve, sleep_for_ms));
        }
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
     * Retry a failed, cancelled, or expired job.
     */
    async retry(job_id: string, options?: RequestOptions): Promise<Job> {
        return this._fetchJson(ZJob, {
            url: `/v1/jobs/${job_id}/retry`,
            method: "POST",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List jobs with pagination and optional filtering.
     */
    async list({
        before,
        after,
        limit = 20,
        order = "desc",
        id,
        status,
        endpoint,
        source,
        project_id,
        workflow_id,
        workflow_node_id,
        model,
        filename_regex,
        filename_contains,
        document_type,
        from_date,
        to_date,
        metadata,
        include_request,
        include_response,
    }: {
        before?: string;
        after?: string;
        limit?: number;
        order?: JobListOrder;
        id?: string;
        status?: JobStatus;
        endpoint?: SupportedEndpoint;
        source?: JobListSource;
        project_id?: string;
        workflow_id?: string;
        workflow_node_id?: string;
        model?: string;
        filename_regex?: string;
        filename_contains?: string;
        document_type?: string[];
        from_date?: string;
        to_date?: string;
        metadata?: Record<string, string>;
        include_request?: boolean;
        include_response?: boolean;
    } = {}, options?: RequestOptions): Promise<JobListResponse> {
        const params: Record<string, any> = {
            before,
            after,
            limit,
            order,
            id,
            status,
            endpoint,
            source,
            project_id,
            workflow_id,
            workflow_node_id,
            model,
            filename_regex,
            filename_contains,
            document_type,
            from_date,
            to_date,
            metadata: metadata ? JSON.stringify(metadata) : undefined,
            include_request,
            include_response,
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

export { Job, JobListOrder, JobListResponse, JobListSource, JobStatus, SupportedEndpoint, JobResponse, JobError };
export { JobRetrieveOptions, JobWaitForCompletionOptions };
