import { AbstractClient, CompositionClient } from "../client.js";
import APIModels from "./models/client";
import APIDocuments from "./documents/client";
import APISchemas from "./schemas/client";
import APIProjects from "./projects/client";
import APIExtractions from "./extractions/client";

export default class APIV1 extends CompositionClient {
    constructor(client: AbstractClient) {
        super(client);
    }
    models = new APIModels(this);
    documents = new APIDocuments(this);
    schemas = new APISchemas(this);
    projects = new APIProjects(this);
    extractions = new APIExtractions(this);
}
