import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIFinetuningSub from "./finetuning/client";
import APIModelCardsForPickerSub from "./modelCardsForPicker/client";
import APIOpenaiTierSub from "./openaiTier/client";
import { ZModelsResponse, ModelsResponse } from "@/types";

export default class APIModels extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  finetuning = new APIFinetuningSub(this._client);
  modelCardsForPicker = new APIModelCardsForPickerSub(this._client);
  openaiTier = new APIOpenaiTierSub(this._client);

  async get({ supportsFinetuning, supportsImage, includeFinetunedModels }: { supportsFinetuning?: boolean, supportsImage?: boolean, includeFinetunedModels?: boolean } = {}): Promise<ModelsResponse> {
    let res = await this._fetch({
      url: `/v1/models`,
      method: "GET",
      params: { "supports_finetuning": supportsFinetuning, "supports_image": supportsImage, "include_finetuned_models": includeFinetunedModels },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZModelsResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
