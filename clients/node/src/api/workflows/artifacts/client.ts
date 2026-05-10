import { CompositionClient, RequestOptions } from "../../../client.js";
import {
    StepArtifactRef,
    WorkflowArtifact,
    ZWorkflowArtifact,
} from "../../../types.js";
import * as z from "zod";

/**
 * Workflow Artifacts API client for dereferencing step artifact refs.
 *
 * Usage: `client.workflows.artifacts.get(step.artifact)`
 */
export default class APIWorkflowArtifacts extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Dereference a workflow step artifact ref into its persisted record.
     *
     * @example
     * ```typescript
     * const step = await client.workflows.runs.steps.get(run.id, "hil-1");
     * if (step.artifact) {
     *   const artifact = await client.workflows.artifacts.get(step.artifact);
     *   console.log(artifact.operation, artifact.evaluations);
     * }
     * ```
     */
    async get(
        artifact: StepArtifactRef,
        options?: RequestOptions
    ): Promise<WorkflowArtifact>;
    async get(
        operation: string,
        id: string,
        options?: RequestOptions
    ): Promise<WorkflowArtifact>;
    async get(
        operationOrArtifact: string | StepArtifactRef,
        idOrOptions?: string | RequestOptions,
        maybeOptions?: RequestOptions
    ): Promise<WorkflowArtifact> {
        const operation =
            typeof operationOrArtifact === "string"
                ? operationOrArtifact
                : operationOrArtifact.operation;
        const id =
            typeof operationOrArtifact === "string"
                ? idOrOptions
                : operationOrArtifact.id;
        const options =
            typeof operationOrArtifact === "string"
                ? maybeOptions
                : (idOrOptions as RequestOptions | undefined);

        if (typeof id !== "string" || id.length === 0) {
            throw new TypeError("artifact id is required");
        }
        return this._fetchJson(ZWorkflowArtifact, {
            url: `/workflows/artifacts/${operation}/${id}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List dereferenced artifacts produced by one workflow run.
     */
    async list(
        {
            runId,
            operation,
            blockId,
        }: {
            runId: string;
            operation?: string;
            blockId?: string;
        },
        options?: RequestOptions
    ): Promise<WorkflowArtifact[]> {
        const params = Object.fromEntries(
            Object.entries({
                run_id: runId,
                operation,
                block_id: blockId,
                ...(options?.params || {}),
            }).filter(([, value]) => value !== undefined)
        );

        return this._fetchJson(z.array(ZWorkflowArtifact), {
            url: "/workflows/artifacts",
            method: "GET",
            params,
            headers: options?.headers,
        });
    }
}
