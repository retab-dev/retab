import { AbstractClient, CompositionClient } from '@/client';
import APISampleDocumentSub from "./sampleDocument/client";
import { TemplateSchema, UpdateTemplateRequest, TemplateSchema } from "@/types";

export default class APITemplateId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  sampleDocument = new APISampleDocumentSub(this._client);

  async get(templateId: string): Promise<TemplateSchema> {
    return this._fetch({
      url: `/v1/schemas/default_templates/${templateId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async delete(templateId: string): Promise<object> {
    return this._fetch({
      url: `/v1/schemas/default_templates/${templateId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async put(templateId: string, { ...body }: UpdateTemplateRequest): Promise<TemplateSchema> {
    return this._fetch({
      url: `/v1/schemas/default_templates/${templateId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
