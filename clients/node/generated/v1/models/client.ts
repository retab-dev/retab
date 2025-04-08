import { AbstractClient, CompositionClient } from '@/client';
import APIFinetuning from "./finetuning/client";
import APIModelCards from "./modelCards/client";
import { ModelsResponse } from "@/types";

export default class APIModels extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  finetuning = new APIFinetuning(this);
  modelCards = new APIModelCards(this);

  async get({ finetuning }: { finetuning?: boolean }): Promise<ModelsResponse> {
    return this._fetch({
      url: `/v1/models`,
      method: "GET",
      params: { "finetuning": finetuning },
      headers: {  },
    });
  }
  
}
