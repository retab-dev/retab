import { AbstractClient, CompositionClient } from '@/client';
import APIFinetuningSub from "./finetuning/client";
import APIModelCardsSub from "./modelCards/client";
import { ModelsResponse } from "@/types";

export default class APIModels extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  finetuning = new APIFinetuningSub(this._client);
  modelCards = new APIModelCardsSub(this._client);

  async get({ finetuning }: { finetuning?: boolean } = {}): Promise<ModelsResponse> {
    return this._fetch({
      url: `/v1/models`,
      method: "GET",
      params: { "finetuning": finetuning },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
