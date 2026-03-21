import { AbstractClient, CompositionClient } from "../client.js";
import APIDocuments from "./documents/client";
import APISchemas from "./schemas/client";
import APIProjects from "./projects/client";
import APIExtractions from "./extractions/client";
import APIWorkflows from "./workflows/client";
import APIEdit from "./edit/client";
import APIFiles from "./files/client";
import APIJobs from "./jobs/client";
import APIEvals from "./evals/client";
import APIModels from "./models/client";

export default class APIV1 extends CompositionClient {
    constructor(client: AbstractClient) {
        super(client);
    }
    files = new APIFiles(this);
    documents = new APIDocuments(this);
    schemas = new APISchemas(this);
    projects = new APIProjects(this);
    evals = new APIEvals(this);
    extractions = new APIExtractions(this);
    workflows = new APIWorkflows(this);
    edit = new APIEdit(this);
    jobs = new APIJobs(this);
    models = new APIModels(this);
}
