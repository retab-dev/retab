import { AbstractClient, CompositionClient } from '@/client';
import APITemplatesSub from "./templates/client";
import APIDefaultTemplatesSub from "./defaultTemplates/client";
import APIHistoryTemplatesSub from "./historyTemplates/client";
import APIPromptifySub from "./promptify/client";
import APIGenerateSub from "./generate/client";
import APISchemaSub from "./schema/client";
import APIExtractionCountSub from "./extractionCount/client";
import APIDataIdSub from "./dataId/client";
import APISchemaIdSub from "./schemaId/client";
import APISystemPromptSub from "./systemPrompt/client";
import { ListSchemas } from "@/types";

export default class APISchemas extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  templates = new APITemplatesSub(this._client);
  defaultTemplates = new APIDefaultTemplatesSub(this._client);
  historyTemplates = new APIHistoryTemplatesSub(this._client);
  promptify = new APIPromptifySub(this._client);
  generate = new APIGenerateSub(this._client);
  schema = new APISchemaSub(this._client);
  extractionCount = new APIExtractionCountSub(this._client);
  dataId = new APIDataIdSub(this._client);
  schemaId = new APISchemaIdSub(this._client);
  systemPrompt = new APISystemPromptSub(this._client);

  async get({ before, after, limit, order, name, sortBy }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", name?: string | null, sortBy?: string } = {}): Promise<ListSchemas> {
    return this._fetch({
      url: `/v1/schemas`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "name": name, "sort_by": sortBy },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
