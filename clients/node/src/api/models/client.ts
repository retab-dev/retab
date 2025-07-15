import { CompositionClient } from "../../client";
import { Model, ZModel } from "../../types";

export default class APIModels extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    
    async list(params: {
        supports_finetuning?: boolean,
        supports_image?: boolean,
        include_finetuned_models?: boolean,
    }): Promise<Model[]> {
        let response = await this._fetch({
            url: "/api/v1/models",
            method: "GET",
            params,
        });
        return ZModel.array().parse(await response.json());
    }
}
