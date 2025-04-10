import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APISampleDocumentSub from "./sampleDocument/client";
import { TemplateSchema, UpdateTemplateRequest, TemplateSchema } from "@/types";

export default class APITemplateId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  sampleDocument = new APISampleDocumentSub(this._client);

  async get(templateId: string): Promise<TemplateSchema> {
    let res = await this._fetch({
      url: `/v1/schemas/templates/${templateId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async delete(templateId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/schemas/templates/${templateId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async put(templateId: string, { ...body }: UpdateTemplateRequest): Promise<TemplateSchema> {
    let res = await this._fetch({
      url: `/v1/schemas/templates/${templateId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
