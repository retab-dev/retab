import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIAutomationsSub from "./automations/client";
import { Branding, BrandingUpdateRequest, Branding } from "@/types";

export default class APIBranding extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  automations = new APIAutomationsSub(this._client);

  async get(): Promise<Branding> {
    let res = await this._fetch({
      url: `/v1/branding`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async post({ ...body }: BrandingUpdateRequest): Promise<Branding> {
    let res = await this._fetch({
      url: `/v1/branding`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async delete(): Promise<object> {
    let res = await this._fetch({
      url: `/v1/branding`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
