import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZConsensusDictRequest, ConsensusDictRequest, ZReconciliationResponse, ReconciliationResponse } from "@/types";

export default class APIReconcile extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: ConsensusDictRequest): Promise<ReconciliationResponse> {
    let res = await this._fetch({
      url: `/v1/consensus/reconcile`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZReconciliationResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
