import { AbstractClient, CompositionClient } from '@/client';
import APITemplateId from "./templateId/client";
import { ListTemplates, CubeServerServicesV1SchemasDefaultTemplatesRoutesCreateTemplateRequest, TemplateSchema } from "@/types";

export default class APIDefaultTemplates extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  templateId = new APITemplateId(this);

  async get({ before, after, limit, order, name, id, dataId, sortBy }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", name?: string | null, id?: string | null, dataId?: string | null, sortBy?: string }): Promise<ListTemplates> {
    return this._fetch({
      url: `/v1/schemas/default_templates`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "name": name, "id": id, "data_id": dataId, "sort_by": sortBy },
      headers: {  },
    });
  }
  
  async post({ ...body }: CubeServerServicesV1SchemasDefaultTemplatesRoutesCreateTemplateRequest): Promise<TemplateSchema> {
    return this._fetch({
      url: `/v1/schemas/default_templates`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
