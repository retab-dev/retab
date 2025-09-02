import { CompositionClient, RequestOptions } from "../../client.js";
import { ZModel } from "../../types.js";
import * as z from "zod";
export default class APIModels extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    async list(params?: {
        supports_finetuning?: boolean,
        supports_image?: boolean,
        include_finetuned_models?: boolean,
    }, options?: RequestOptions) {
        return this._fetchJson(z.object({ data: z.array(ZModel) }), {
            url: "/v1/models",
            method: "GET",
            params: { ...(params || {}), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }
}
