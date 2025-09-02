import { CompositionClient, RequestOptions } from "../../client.js";
import { GenerateSchemaRequest, ZGenerateSchemaRequest, ZSchema } from "../../types.js";

export default class APISchemas extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    async generate(params: GenerateSchemaRequest, options?: RequestOptions) {
        return this._fetchJson(ZSchema, {
            url: "/v1/schemas/generate",
            method: "POST",
            body: { ...(await ZGenerateSchemaRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}
