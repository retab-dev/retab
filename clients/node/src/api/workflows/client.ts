import { CompositionClient } from "../../client.js";
import APIWorkflowRuns from "./runs/client.js";

/**
 * Workflows API client for workflow operations.
 *
 * Sub-clients:
 * - runs: Workflow run operations (create, get)
 */
export default class APIWorkflows extends CompositionClient {
    public runs: APIWorkflowRuns;

    constructor(client: CompositionClient) {
        super(client);
        this.runs = new APIWorkflowRuns(this);
    }
}
