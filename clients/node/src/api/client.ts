import { AbstractClient, CompositionClient } from "@/client";
import APIModels from "./models/client";
import APIConsensus from "./consensus/client";
import APIDocuments from "./documents/client";
import APISchemas from "./schemas/client";
import APIDeployments from "./deployments/client";

export default class APIV1 extends CompositionClient {
    constructor(client: AbstractClient) {
        super(client);
    }
    
    models = new APIModels(this);
    consensus = new APIConsensus(this);
    documents = new APIDocuments(this);
    schemas = new APISchemas(this);
    deployments = new APIDeployments(this);
}
