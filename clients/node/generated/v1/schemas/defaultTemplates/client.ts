import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APITemplateIdSub from "./templateId/client";
import { ZListTemplates, ListTemplates, ZMainServerServicesV1SchemasDefaultTemplatesRoutesCreateTemplateRequest, MainServerServicesV1SchemasDefaultTemplatesRoutesCreateTemplateRequest, ZTemplateSchema, TemplateSchema } from "@/types";

export default class APIDefaultTemplates extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  templateId = new APITemplateIdSub(this._client);

  async get({ before, after, limit, order, name, id, dataId, sortBy }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", name?: string | null, id?: string | null, dataId?: string | null, sortBy?: string } = {}): Promise<ListTemplates> {
    let res = await this._fetch({
      url: `/v1/schemas/default_templates`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "name": name, "id": id, "data_id": dataId, "sort_by": sortBy },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZListTemplates.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async post({ ...body }: MainServerServicesV1SchemasDefaultTemplatesRoutesCreateTemplateRequest): Promise<TemplateSchema> {
    let res = await this._fetch({
      url: `/v1/schemas/default_templates`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZTemplateSchema.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
