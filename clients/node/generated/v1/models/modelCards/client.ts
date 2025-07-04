import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ModelCardsResponse } from "@/types";

export default class APIModelCards extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ finetuning }: { finetuning?: boolean } = {}): Promise<ModelCardsResponse> {
    let res = await this._fetch({
      url: `/v1/models/model_cards`,
      method: "GET",
      params: { "finetuning": finetuning },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
