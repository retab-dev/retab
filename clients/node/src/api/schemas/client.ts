import { CompositionClient } from "../../client.js";
import { GenerateSchemaRequest, ZGenerateSchemaRequest, ZSchema } from "../../types.js";

export default class APISchemas extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    async generate(params: GenerateSchemaRequest) {
        return this._fetchJson(ZSchema, {
            url: "/v1/schemas/generate",
            method: "POST",
            body: await ZGenerateSchemaRequest.parseAsync(params),
        });
    }
}
