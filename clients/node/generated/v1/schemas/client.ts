import { AbstractClient, CompositionClient } from '@/client';
import APITemplates from "./templates/client";
import APIDefaultTemplates from "./defaultTemplates/client";
import APIHistoryTemplates from "./historyTemplates/client";
import APIPromptify from "./promptify/client";
import APIGenerate from "./generate/client";
import APISchema from "./schema/client";
import APIExtractionCount from "./extractionCount/client";
import APIDataId from "./dataId/client";
import APISchemaId from "./schemaId/client";
import APISystemPrompt from "./systemPrompt/client";
import { ListSchemas } from "@/types";

export default class APISchemas extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  templates = new APITemplates(this);
  defaultTemplates = new APIDefaultTemplates(this);
  historyTemplates = new APIHistoryTemplates(this);
  promptify = new APIPromptify(this);
  generate = new APIGenerate(this);
  schema = new APISchema(this);
  extractionCount = new APIExtractionCount(this);
  dataId = new APIDataId(this);
  schemaId = new APISchemaId(this);
  systemPrompt = new APISystemPrompt(this);

  async get({ before, after, limit, order, name, sortBy }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", name?: string | null, sortBy?: string }): Promise<ListSchemas> {
    return this._fetch({
      url: `/v1/schemas`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "name": name, "sort_by": sortBy },
      headers: {  },
    });
  }
  
}
