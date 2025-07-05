import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { FinetunedModel } from "@/types";

export default class APIModelId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(modelId: string): Promise<FinetunedModel> {
    let res = await this._fetch({
      url: `/v1/models/finetuning/${modelId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
