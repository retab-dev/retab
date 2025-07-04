import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ModelCardsResponse } from "@/types";

export default class APIModelCardsForPicker extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ supportsFinetuning, includeFinetunedModels, supportsSchemaGeneration, supportsImage }: { supportsFinetuning?: boolean, includeFinetunedModels?: boolean, supportsSchemaGeneration?: boolean, supportsImage?: boolean } = {}): Promise<ModelCardsResponse> {
    let res = await this._fetch({
      url: `/v1/models/model_cards_for_picker`,
      method: "GET",
      params: { "supports_finetuning": supportsFinetuning, "include_finetuned_models": includeFinetunedModels, "supports_schema_generation": supportsSchemaGeneration, "supports_image": supportsImage },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
