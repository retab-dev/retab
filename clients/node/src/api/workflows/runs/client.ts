import { CompositionClient, RequestOptions } from "../../../client.js";
import { MIMEDataInput, ZMIMEData, WorkflowRun, ZWorkflowRun } from "../../../types.js";

/**
 * Workflow Runs API client for managing workflow executions.
 */
export default class APIWorkflowRuns extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
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
        // Build the request body
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const body: Record<string, any> = {};

        // Convert each document to MIMEData format expected by backend
        if (documents) {
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
            body.documents = documentsPayload;
        }

        // Add JSON inputs directly
        if (jsonInputs) {
            body.json_inputs = jsonInputs;
        }

        // Add text inputs directly
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
}
