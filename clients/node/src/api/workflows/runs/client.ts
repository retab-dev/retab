import { CompositionClient, RequestOptions } from "../../../client.js";
import {
    MIMEDataInput,
    ZMIMEData,
    WorkflowRun,
    ZWorkflowRun,
    PaginatedList,
    ZPaginatedList,
    CancelWorkflowResponse,
    ZCancelWorkflowResponse,
    ResumeWorkflowResponse,
    ZResumeWorkflowResponse,
} from "../../../types.js";
import APIWorkflowRunSteps from "./steps/client.js";

const TERMINAL_STATUSES = new Set(["completed", "error", "cancelled"]);

function sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Workflow Runs API client for managing workflow executions.
 *
 * Sub-clients:
 * - steps: Step output operations (get, getBatch)
 */
export default class APIWorkflowRuns extends CompositionClient {
    public steps: APIWorkflowRunSteps;

    constructor(client: CompositionClient) {
        super(client);
        this.steps = new APIWorkflowRunSteps(this);
    }

    /**
     * Run a workflow with the provided inputs.
     *
     * This creates a workflow run and starts execution in the background.
     * The returned WorkflowRun will have status "running" - use get()
     * to check for updates on the run status.
     *
     * @param workflowId - The ID of the workflow to run
     * @param documents - Mapping of start node IDs to their input documents
     * @param jsonInputs - Mapping of start_json node IDs to their input JSON data
     * @param textInputs - Mapping of start_text node IDs to their input text
     * @param options - Optional request options
     * @returns The created workflow run with status "running"
     *
     * @example
     * ```typescript
     * const run = await client.workflows.runs.create({
     *     workflowId: "wf_abc123",
     *     documents: {
     *         "start-node-1": "./invoice.pdf",
     *     },
     *     jsonInputs: {
     *         "json-node-1": { key: "value" },
     *     },
     *     textInputs: {
     *         "text-node-1": "Hello, world!",
     *     }
     * });
     * console.log(`Run started: ${run.id}, status: ${run.status}`);
     * ```
     */
    async create(
        {
            workflowId,
            documents,
            jsonInputs,
            textInputs,
        }: {
            workflowId: string;
            documents?: Record<string, MIMEDataInput>;
            jsonInputs?: Record<string, Record<string, unknown>>;
            textInputs?: Record<string, string>;
        },
        options?: RequestOptions
    ): Promise<WorkflowRun> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {};

        if (documents) {
            const documentsPayload: Record<string, { filename: string; content: string; mime_type: string }> = {};

            for (const [nodeId, document] of Object.entries(documents)) {
                const parsedDocument = await ZMIMEData.parseAsync(document);
                const content = parsedDocument.url.split(",")[1];
                const mimeType = parsedDocument.url.split(";")[0].split(":")[1];

                documentsPayload[nodeId] = {
                    filename: parsedDocument.filename,
                    content: content,
                    mime_type: mimeType,
                };
            }
            body.documents = documentsPayload;
        }

        if (jsonInputs) {
            body.json_inputs = jsonInputs;
        }

        if (textInputs) {
            body.text_inputs = textInputs;
        }

        return this._fetchJson(ZWorkflowRun, {
            url: `/v1/workflows/${workflowId}/run`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get a workflow run by ID.
     *
     * @param runId - The ID of the workflow run to retrieve
     * @param options - Optional request options
     * @returns The workflow run
     *
     * @example
     * ```typescript
     * const run = await client.workflows.runs.get("run_abc123");
     * console.log(`Run status: ${run.status}`);
     * ```
     */
    async get(runId: string, options?: RequestOptions): Promise<WorkflowRun> {
        return this._fetchJson(ZWorkflowRun, {
            url: `/v1/workflows/runs/${runId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List workflow runs with filtering and pagination.
     *
     * @example
     * ```typescript
     * const runs = await client.workflows.runs.list({
     *     workflowId: "wf_abc123",
     *     status: "completed",
     *     limit: 10,
     * });
     * console.log(`Found ${runs.data.length} runs`);
     * ```
     */
    async list(
        {
            workflowId,
            status,
            statuses,
            excludeStatus,
            triggerType,
            triggerTypes,
            fromDate,
            toDate,
            minCost,
            maxCost,
            minDuration,
            maxDuration,
            search,
            sortBy,
            fields,
            before,
            after,
            limit = 20,
            order = "desc",
        }: {
            workflowId?: string;
            status?: "pending" | "running" | "completed" | "error" | "waiting_for_human" | "cancelled";
            statuses?: string;
            excludeStatus?: "pending" | "running" | "completed" | "error" | "waiting_for_human" | "cancelled";
            triggerType?: "manual" | "api" | "schedule" | "webhook" | "restart";
            triggerTypes?: string;
            fromDate?: string;
            toDate?: string;
            minCost?: number;
            maxCost?: number;
            minDuration?: number;
            maxDuration?: number;
            search?: string;
            sortBy?: string;
            fields?: string;
            before?: string;
            after?: string;
            limit?: number;
            order?: "asc" | "desc";
        } = {},
        options?: RequestOptions
    ): Promise<PaginatedList> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const params: Record<string, any> = {
            workflow_id: workflowId,
            status,
            statuses,
            exclude_status: excludeStatus,
            trigger_type: triggerType,
            trigger_types: triggerTypes,
            from_date: fromDate,
            to_date: toDate,
            min_cost: minCost,
            max_cost: maxCost,
            min_duration: minDuration,
            max_duration: maxDuration,
            search,
            sort_by: sortBy,
            fields,
            before,
            after,
            limit,
            order,
        };

        const cleanParams = Object.fromEntries(
            Object.entries(params).filter(([_, v]) => v !== undefined)
        );

        return this._fetchJson(ZPaginatedList, {
            url: "/v1/workflows/runs",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Delete a workflow run and its associated step data.
     *
     * @param runId - The ID of the workflow run to delete
     * @param options - Optional request options
     *
     * @example
     * ```typescript
     * await client.workflows.runs.delete("run_abc123");
     * ```
     */
    async delete(runId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/v1/workflows/runs/${runId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Cancel a running or pending workflow run.
     *
     * @param runId - The ID of the workflow run to cancel
     * @param commandId - Optional idempotency key for deduplicating cancel commands
     * @param options - Optional request options
     * @returns The updated run with cancellation status
     *
     * @example
     * ```typescript
     * const result = await client.workflows.runs.cancel("run_abc123");
     * console.log(`Cancellation: ${result.cancellation_status}`);
     * ```
     */
    async cancel(
        runId: string,
        { commandId }: { commandId?: string } = {},
        options?: RequestOptions
    ): Promise<CancelWorkflowResponse> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {};
        if (commandId !== undefined) body.command_id = commandId;

        return this._fetchJson(ZCancelWorkflowResponse, {
            url: `/v1/workflows/runs/${runId}/cancel`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Restart a completed or failed workflow run with the same inputs.
     *
     * @param runId - The ID of the workflow run to restart
     * @param commandId - Optional idempotency key for deduplicating restart commands
     * @param options - Optional request options
     * @returns The new workflow run
     *
     * @example
     * ```typescript
     * const newRun = await client.workflows.runs.restart("run_abc123");
     * console.log(`New run: ${newRun.id}`);
     * ```
     */
    async restart(
        runId: string,
        { commandId }: { commandId?: string } = {},
        options?: RequestOptions
    ): Promise<WorkflowRun> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {};
        if (commandId !== undefined) body.command_id = commandId;

        return this._fetchJson(ZWorkflowRun, {
            url: `/v1/workflows/runs/${runId}/restart`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Resume a workflow run after human-in-the-loop (HIL) review.
     *
     * @param runId - The ID of the workflow run to resume
     * @param nodeId - The ID of the HIL node being approved/rejected
     * @param approved - Whether the human approved the data
     * @param modifiedData - Optional modified data if the human made changes
     * @param commandId - Optional idempotency key for deduplicating resume commands
     * @param options - Optional request options
     * @returns The resume response with status and queue information
     *
     * @example
     * ```typescript
     * const result = await client.workflows.runs.resume("run_abc123", {
     *     nodeId: "hil-node-1",
     *     approved: true,
     *     modifiedData: { field: "corrected value" },
     * });
     * console.log(`Resume status: ${result.resume_status}`);
     * ```
     */
    async resume(
        runId: string,
        {
            nodeId,
            approved,
            modifiedData,
            commandId,
        }: {
            nodeId: string;
            approved: boolean;
            modifiedData?: Record<string, unknown> | null;
            commandId?: string;
        },
        options?: RequestOptions
    ): Promise<ResumeWorkflowResponse> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {
            node_id: nodeId,
            approved,
        };
        if (modifiedData !== undefined) body.modified_data = modifiedData;
        if (commandId !== undefined) body.command_id = commandId;

        return this._fetchJson(ZResumeWorkflowResponse, {
            url: `/v1/workflows/runs/${runId}/resume`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Poll a workflow run until it reaches a terminal state.
     *
     * Terminal states: "completed", "error", "cancelled".
     *
     * @param runId - The ID of the workflow run to wait for
     * @param pollIntervalMs - Milliseconds between polls (default: 2000)
     * @param timeoutMs - Maximum time to wait in milliseconds (default: 600000 = 10 minutes)
     * @returns The workflow run in a terminal state
     * @throws Error if the run doesn't complete within the timeout
     *
     * @example
     * ```typescript
     * const run = await client.workflows.runs.waitForCompletion("run_abc123", {
     *     pollIntervalMs: 1000,
     *     timeoutMs: 300000,
     * });
     * console.log(`Final status: ${run.status}`);
     * ```
     */
    async waitForCompletion(
        runId: string,
        {
            pollIntervalMs = 2000,
            timeoutMs = 600000,
        }: {
            pollIntervalMs?: number;
            timeoutMs?: number;
        } = {}
    ): Promise<WorkflowRun> {
        if (pollIntervalMs <= 0) throw new Error("pollIntervalMs must be positive");
        if (timeoutMs <= 0) throw new Error("timeoutMs must be positive");

        const deadline = Date.now() + timeoutMs;

        while (true) {
            const run = await this.get(runId);

            if (TERMINAL_STATUSES.has(run.status)) {
                return run;
            }

            if (Date.now() >= deadline) {
                throw new Error(
                    `Workflow run ${runId} did not complete within ${timeoutMs}ms (last status: ${run.status})`
                );
            }

            await sleep(pollIntervalMs);
        }
    }

    /**
     * Create a workflow run and wait for it to complete.
     *
     * Convenience method that combines create() and waitForCompletion().
     *
     * @param workflowId - The ID of the workflow to run
     * @param documents - Mapping of start node IDs to their input documents
     * @param jsonInputs - Mapping of start_json node IDs to their input JSON data
     * @param textInputs - Mapping of start_text node IDs to their input text
     * @param pollIntervalMs - Milliseconds between polls (default: 2000)
     * @param timeoutMs - Maximum time to wait in milliseconds (default: 600000)
     * @param options - Optional request options
     * @returns The workflow run in a terminal state
     *
     * @example
     * ```typescript
     * const run = await client.workflows.runs.createAndWait({
     *     workflowId: "wf_abc123",
     *     documents: { "start-node-1": "./invoice.pdf" },
     *     timeoutMs: 120000,
     * });
     * if (run.status === "completed") {
     *     console.log("Outputs:", run.final_outputs);
     * }
     * ```
     */
    async createAndWait(
        {
            workflowId,
            documents,
            jsonInputs,
            textInputs,
            pollIntervalMs = 2000,
            timeoutMs = 600000,
        }: {
            workflowId: string;
            documents?: Record<string, MIMEDataInput>;
            jsonInputs?: Record<string, Record<string, unknown>>;
            textInputs?: Record<string, string>;
            pollIntervalMs?: number;
            timeoutMs?: number;
        },
        options?: RequestOptions
    ): Promise<WorkflowRun> {
        const run = await this.create(
            { workflowId, documents, jsonInputs, textInputs },
            options
        );
        return this.waitForCompletion(run.id, { pollIntervalMs, timeoutMs });
    }
}
