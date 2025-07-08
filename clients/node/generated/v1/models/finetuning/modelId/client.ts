import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZFinetunedModel, FinetunedModel } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return ZFinetunedModel.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
