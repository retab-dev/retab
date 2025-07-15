import { AbstractClient, CompositionClient } from "@/client";
import APIModels from "./models/client";
import APIConsensus from "./consensus/client";

export default class APIV1 extends CompositionClient {
    constructor(client: AbstractClient) {
        super(client);
    }
    
    models = new APIModels(this);
    consensus = new APIConsensus(this);
}
