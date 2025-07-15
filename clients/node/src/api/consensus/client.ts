import { CompositionClient } from "@/client";
import { ReconciliationRequest, ReconciliationResponse, ZReconciliationResponse } from "@/types";

export default class APIConsensus extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    async reconcile(params: ReconciliationRequest): Promise<ReconciliationResponse> {
        return await this._fetchJson(ZReconciliationResponse, {
            url: "/v1/consensus/reconcile",
            method: "POST",
            body: params,
        });
    }
}
