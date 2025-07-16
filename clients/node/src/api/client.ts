import { AbstractClient, CompositionClient } from "@/client";
import APIModels from "./models/client";
import APIConsensus from "./consensus/client";
import APIDocuments from "./documents/client";

export default class APIV1 extends CompositionClient {
    constructor(client: AbstractClient) {
        super(client);
    }
    
    models = new APIModels(this);
    consensus = new APIConsensus(this);
    documents = new APIDocuments(this);
}
