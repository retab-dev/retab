import { AbstractClient, CompositionClient } from '@/client';
import { FinetunedModel } from "@/types";

export default class APIModel extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(model: string): Promise<FinetunedModel> {
    return this._fetch({
      url: `/v1/finetuned_models/${model}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
