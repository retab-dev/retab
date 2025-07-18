import { CompositionClient } from "@/client";
import { GenerateSchemaRequest, ZGenerateSchemaRequest, ZSchema } from "@/types";

export default class APIModels extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    
    async generate(params: GenerateSchemaRequest) {
        return await this._fetchJson(ZSchema, {
            url: "/v1/schemas/generate",
            method: "POST",
            body: await ZGenerateSchemaRequest.parseAsync(params),
        });
    }
}
