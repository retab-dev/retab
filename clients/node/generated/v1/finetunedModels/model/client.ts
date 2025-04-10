import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { FinetunedModel } from "@/types";

export default class APIModel extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(model: string): Promise<FinetunedModel> {
    let res = await this._fetch({
      url: `/v1/finetuned_models/${model}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
