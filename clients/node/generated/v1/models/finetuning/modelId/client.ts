import { AbstractClient, CompositionClient } from '@/client';
import { FinetunedModel } from "@/types";

export default class APIModelId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(modelId: string): Promise<FinetunedModel> {
    return this._fetch({
      url: `/v1/models/finetuning/${modelId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
