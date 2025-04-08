import { AbstractClient, CompositionClient } from '@/client';
import APITemplateId from "./templateId/client";
import { ListTemplates, CubeServerServicesV1SchemasTemplatesRoutesCreateTemplateRequest, TemplateSchema } from "@/types";

export default class APITemplates extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  templateId = new APITemplateId(this);

  async get({ before, after, limit, order, name, id, dataId, sortBy }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", name?: string | null, id?: string | null, dataId?: string | null, sortBy?: string }): Promise<ListTemplates> {
    return this._fetch({
      url: `/v1/schemas/templates`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "name": name, "id": id, "data_id": dataId, "sort_by": sortBy },
      headers: {  },
    });
  }
  
  async post({ ...body }: CubeServerServicesV1SchemasTemplatesRoutesCreateTemplateRequest): Promise<TemplateSchema> {
    return this._fetch({
      url: `/v1/schemas/templates`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
