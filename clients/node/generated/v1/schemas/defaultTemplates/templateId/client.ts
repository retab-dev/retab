import { AbstractClient, CompositionClient } from '@/client';
import APISampleDocument from "./sampleDocument/client";
import { TemplateSchema, UpdateTemplateRequest, TemplateSchema } from "@/types";

export default class APITemplateId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  sampleDocument = new APISampleDocument(this);

  async get(templateId: string): Promise<TemplateSchema> {
    return this._fetch({
      url: `/v1/schemas/default_templates/${templateId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
  async delete(templateId: string): Promise<object> {
    return this._fetch({
      url: `/v1/schemas/default_templates/${templateId}`,
      method: "DELETE",
      params: {  },
      headers: {  },
    });
  }
  
  async put(templateId: string, { ...body }: UpdateTemplateRequest): Promise<TemplateSchema> {
    return this._fetch({
      url: `/v1/schemas/default_templates/${templateId}`,
      method: "PUT",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
