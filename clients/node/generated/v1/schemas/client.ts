import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APITemplatesSub from "./templates/client";
import APIDefaultTemplatesSub from "./defaultTemplates/client";
import APIHistoryTemplatesSub from "./historyTemplates/client";
import APIGenerateSub from "./generate/client";
import APIExtractionCountSub from "./extractionCount/client";
import APISystemPromptEndpointSub from "./systemPromptEndpoint/client";
import APIEvaluateSub from "./evaluate/client";
import APIEnhanceSub from "./enhance/client";
import APICachePreloadSub from "./cachePreload/client";

export default class APISchemas extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  templates = new APITemplatesSub(this._client);
  defaultTemplates = new APIDefaultTemplatesSub(this._client);
  historyTemplates = new APIHistoryTemplatesSub(this._client);
  generate = new APIGenerateSub(this._client);
  extractionCount = new APIExtractionCountSub(this._client);
  systemPromptEndpoint = new APISystemPromptEndpointSub(this._client);
  evaluate = new APIEvaluateSub(this._client);
  enhance = new APIEnhanceSub(this._client);
  cachePreload = new APICachePreloadSub(this._client);

}
