import { CompositionClient, RequestOptions } from "../../client.js";
import { MIMEDataInput, ZMIMEData, WorkflowRun, ZWorkflowRun } from "../../types.js";

export default class APIWorkflows extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Run a workflow with the provided input documents.
     * 
     * This creates a workflow run and starts execution in the background.
     * The returned WorkflowRun will have status "running" - use getRun() 
     * to check for updates on the run status.
     * 
     * @param workflowId - The ID of the workflow to run
     * @param documents - Mapping of start node IDs to their input documents
     * @param options - Optional request options
     * @returns The created workflow run with status "running"
     * 
     * @example
     * ```typescript
     * const run = await client.workflows.run({
     *     workflowId: "wf_abc123",
     *     documents: {
     *         "start-node-1": "./invoice.pdf",
     *         "start-node-2": Buffer.from(...)
     *     }
     * });
     * console.log(`Run started: ${run.id}, status: ${run.status}`);
     * ```
     */
    async run({
        workflowId,
        documents,
    }: {
        workflowId: string;
        documents: Record<string, MIMEDataInput>;
    }, options?: RequestOptions): Promise<WorkflowRun> {
        // Convert each document to MIMEData format expected by backend
        const documentsPayload: Record<string, { filename: string; content: string; mime_type: string }> = {};

        for (const [nodeId, document] of Object.entries(documents)) {
            const parsedDocument = await ZMIMEData.parseAsync(document);
            // Extract base64 content from data URL
            const content = parsedDocument.url.split(",")[1];
            const mimeType = parsedDocument.url.split(";")[0].split(":")[1];

            documentsPayload[nodeId] = {
                filename: parsedDocument.filename,
                content: content,
                mime_type: mimeType,
            };
        }

        return this._fetchJson(ZWorkflowRun, {
            url: `/v1/workflows/${workflowId}/run`,
            method: "POST",
            body: { documents: documentsPayload, ...(options?.body || {}) },
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
     * const run = await client.workflows.getRun("run_abc123");
     * console.log(`Run status: ${run.status}`);
     * ```
     */
    async getRun(runId: string, options?: RequestOptions): Promise<WorkflowRun> {
        return this._fetchJson(ZWorkflowRun, {
            url: `/v1/workflows/runs/${runId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }
}

