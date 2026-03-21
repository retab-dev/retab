import { CompositionClient, RequestOptions } from "../../client.js";
import { dataArray, Model, ZModel } from "../../types.js";

export default class APIModels extends CompositionClient {
    async list({
        supports_finetuning = false,
        supports_image = false,
        include_finetuned_models = true,
    }: {
        supports_finetuning?: boolean;
        supports_image?: boolean;
        include_finetuned_models?: boolean;
    } = {}, options?: RequestOptions): Promise<Model[]> {
        return this._fetchJson(dataArray(ZModel), {
            url: "/v1/models",
            method: "GET",
            params: {
                supports_finetuning,
                supports_image,
                include_finetuned_models,
                ...(options?.params || {}),
            },
            headers: options?.headers,
        });
    }
}
