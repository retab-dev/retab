import { CompositionClient } from "@/client";
import { ZModel } from "@/types";
import * as z from "zod";
export default class APIModels extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    
    async list(params?: {
        supports_finetuning?: boolean,
        supports_image?: boolean,
        include_finetuned_models?: boolean,
    }) {
        return this._fetchJson(z.object({ data: z.array(ZModel) }), {
            url: "/v1/models",
            method: "GET",
            params: params,
        });
    }
}
