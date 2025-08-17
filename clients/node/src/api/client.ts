import { AbstractClient, CompositionClient } from "../client.js";
import APIModels from "./models/client";
import APIDocuments from "./documents/client";
import APISchemas from "./schemas/client";
import APIDeployments from "./deployments/client";
import APIProjects from "./projects/client";

export default class APIV1 extends CompositionClient {
    constructor(client: AbstractClient) {
        super(client);
    }

    models = new APIModels(this);
    documents = new APIDocuments(this);
    schemas = new APISchemas(this);
    deployments = new APIDeployments(this);
    projects = new APIProjects(this);
}
