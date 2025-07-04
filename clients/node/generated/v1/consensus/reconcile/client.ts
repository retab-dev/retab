import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ConsensusDictRequest, ReconciliationResponse } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
