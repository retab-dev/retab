import * as z from "zod";
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
    WorkflowRunExportResponse,
    ZWorkflowRunExportResponse,
    WorkflowRunStatus,
    WorkflowRunTriggerType,
} from "../../../types.js";
import APIWorkflowRunSteps from "./steps/client.js";

const TERMINAL_STATUSES = new Set(["completed", "error", "cancelled"]);

function normalizeCsvParam(value?: string | string[]): string | undefined {
    if (value === undefined) {
        return undefined;
    }
    return Array.isArray(value) ? value.join(",") : value;
}

function normalizeDateParam(value?: string | Date): string | undefined {
    if (value === undefined) {
        return undefined;
    }
    if (value instanceof Date) {
        return value.toISOString().slice(0, 10);
    }
    return value;
}

function sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Workflow Runs API client for managing workflow executions.
 *
 * Sub-clients:
 * - steps: Step output operations (get, list, getMany, getAll)
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
     * @example
     * ```typescript
     * const run = await client.workflows.runs.create({
     *     workflowId: "wf_abc123",
     *     documents: { "start-node-1": "./invoice.pdf" },
     * });
     * ```
     */
    async create(
        {
            workflowId,
            documents,
            jsonInputs,
        }: {
            workflowId: string;
            documents?: Record<string, MIMEDataInput>;
            jsonInputs?: Record<string, Record<string, unknown>>;
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

        return this._fetchJson(ZWorkflowRun, {
            url: `/workflows/${workflowId}/run`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get a workflow run by ID.
     */
    async get(runId: string, options?: RequestOptions): Promise<WorkflowRun> {
        return this._fetchJson(ZWorkflowRun, {
            url: `/workflows/runs/${runId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List workflow runs with filtering and pagination.
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
            status?: WorkflowRunStatus;
            statuses?: WorkflowRunStatus[] | string;
            excludeStatus?: WorkflowRunStatus;
            triggerType?: WorkflowRunTriggerType;
            triggerTypes?: WorkflowRunTriggerType[] | string;
            fromDate?: string | Date;
            toDate?: string | Date;
            minCost?: number;
            maxCost?: number;
            minDuration?: number;
            maxDuration?: number;
            search?: string;
            sortBy?: string;
            fields?: string[] | string;
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
            statuses: normalizeCsvParam(statuses),
            exclude_status: excludeStatus,
            trigger_type: triggerType,
            trigger_types: normalizeCsvParam(triggerTypes),
            from_date: normalizeDateParam(fromDate),
            to_date: normalizeDateParam(toDate),
            min_cost: minCost,
            max_cost: maxCost,
            min_duration: minDuration,
            max_duration: maxDuration,
            search,
            sort_by: sortBy,
            fields: normalizeCsvParam(fields),
            before,
            after,
            limit,
            order,
        };

        const cleanParams = Object.fromEntries(
            Object.entries(params).filter(([_, v]) => v !== undefined)
        );

        return this._fetchJson(ZPaginatedList, {
            url: "/workflows/runs",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Delete a workflow run and its associated step data.
     */
    async delete(runId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/workflows/runs/${runId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Cancel a running or pending workflow run.
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
            url: `/workflows/runs/${runId}/cancel`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Restart a completed or failed workflow run with the same inputs.
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
            url: `/workflows/runs/${runId}/restart`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Resume a workflow run after human-in-the-loop (HIL) review.
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
            url: `/workflows/runs/${runId}/resume`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Poll a workflow run until it reaches a terminal state.
     *
     * Terminal states: "completed", "error", "cancelled", "waiting_for_human".
     *
     * @param runId - The ID of the workflow run to wait for
     * @param pollIntervalMs - Milliseconds between polls (default: 2000)
     * @param timeoutMs - Maximum time to wait in milliseconds (default: 600000 = 10 minutes)
     * @param onStatus - Optional callback invoked with the WorkflowRun on each poll
     *
     * @example
     * ```typescript
     * const run = await client.workflows.runs.waitForCompletion("run_abc123", {
     *     onStatus: (r) => console.log(`${r.status}...`),
     * });
     * ```
     */
    async waitForCompletion(
        runId: string,
        {
            pollIntervalMs = 2000,
            timeoutMs = 600000,
            onStatus,
        }: {
            pollIntervalMs?: number;
            timeoutMs?: number;
            onStatus?: (run: WorkflowRun) => void | Promise<void>;
        } = {}
    ): Promise<WorkflowRun> {
        if (pollIntervalMs <= 0) throw new Error("pollIntervalMs must be positive");
        if (timeoutMs <= 0) throw new Error("timeoutMs must be positive");

        const deadline = Date.now() + timeoutMs;

        while (true) {
            const run = await this.get(runId);

            if (onStatus) await onStatus(run);

            if (TERMINAL_STATUSES.has(run.status) || run.status === "waiting_for_human") {
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

    async wait_for_completion(
        runId: string,
        {
            poll_interval_ms = 2000,
            timeout_ms = 600000,
            on_status,
        }: {
            poll_interval_ms?: number;
            timeout_ms?: number;
            on_status?: (run: WorkflowRun) => void | Promise<void>;
        } = {}
    ): Promise<WorkflowRun> {
        return this.waitForCompletion(runId, {
            pollIntervalMs: poll_interval_ms,
            timeoutMs: timeout_ms,
            onStatus: on_status,
        });
    }

    /**
     * Create a workflow run and wait for it to complete.
     *
     * @example
     * ```typescript
     * import { raiseForStatus } from "retab";
     *
     * const run = await client.workflows.runs.createAndWait({
     *     workflowId: "wf_abc123",
     *     documents: { "start-node-1": "./invoice.pdf" },
     *     onStatus: (r) => console.log(`${r.status}...`),
     * });
     * raiseForStatus(run);
     * console.log(run.final_outputs);
     * ```
     */
    async createAndWait(
        {
            workflowId,
            documents,
            jsonInputs,
            pollIntervalMs = 2000,
            timeoutMs = 600000,
            onStatus,
        }: {
            workflowId: string;
            documents?: Record<string, MIMEDataInput>;
            jsonInputs?: Record<string, Record<string, unknown>>;
            pollIntervalMs?: number;
            timeoutMs?: number;
            onStatus?: (run: WorkflowRun) => void | Promise<void>;
        },
        options?: RequestOptions
    ): Promise<WorkflowRun> {
        const run = await this.create(
            { workflowId, documents, jsonInputs },
            options
        );
        return this.waitForCompletion(run.id, { pollIntervalMs, timeoutMs, onStatus });
    }

    /**
     * Get the configuration snapshot used for a run.
     */
    async getConfig(runId: string, options?: RequestOptions): Promise<Record<string, unknown>> {
        return this._fetchJson(z.record(z.any()), {
            url: `/workflows/runs/${runId}/config`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get the DAG-ordered execution order for a run.
     */
    async executionOrder(runId: string, options?: RequestOptions): Promise<Record<string, unknown>> {
        return this._fetchJson(z.record(z.any()), {
            url: `/workflows/runs/${runId}/execution-order`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get a signed URL for downloading a document from a run step.
     */
    async getDocumentUrl(
        runId: string,
        nodeId: string,
        options?: RequestOptions
    ): Promise<Record<string, unknown>> {
        return this._fetchJson(z.record(z.any()), {
            url: `/workflows/runs/${runId}/documents/${nodeId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Export run results as structured CSV data.
     *
     * @returns Object with `csv_data` (string), `rows` (number), `columns` (number)
     */
    async export(
        {
            workflowId,
            nodeId,
            exportSource = "outputs",
            selectedRunIds,
            status,
            excludeStatus,
            fromDate,
            toDate,
            triggerTypes,
            preferredColumns,
        }: {
            workflowId: string;
            nodeId: string;
            exportSource?: "outputs" | "inputs";
            selectedRunIds?: string[];
            status?: string;
            excludeStatus?: string;
            fromDate?: string | Date;
            toDate?: string | Date;
            triggerTypes?: WorkflowRunTriggerType[];
            preferredColumns?: string[];
        },
        options?: RequestOptions
    ): Promise<WorkflowRunExportResponse> {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {
            workflow_id: workflowId,
            node_id: nodeId,
            export_source: exportSource,
            preferred_columns: preferredColumns || [],
        };
        if (selectedRunIds !== undefined) body.selected_run_ids = selectedRunIds;
        if (status !== undefined) body.status = status;
        if (excludeStatus !== undefined) body.exclude_status = excludeStatus;
        if (fromDate !== undefined) body.from_date = normalizeDateParam(fromDate);
        if (toDate !== undefined) body.to_date = normalizeDateParam(toDate);
        if (triggerTypes !== undefined) body.trigger_types = triggerTypes;

        return this._fetchJson(ZWorkflowRunExportResponse, {
            url: "/workflows/runs/export_payload",
            method: "POST",
            body: { ...body, ...(options?.body as Record<string, unknown> || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}
