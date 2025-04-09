import { AbstractClient, CompositionClient } from '@/client';
import APIAutomationsSub from "./automations/client";
import { Branding, BrandingUpdateRequest, Branding } from "@/types";

export default class APIBranding extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  automations = new APIAutomationsSub(this._client);

  async get(): Promise<Branding> {
    return this._fetch({
      url: `/v1/branding`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async post({ ...body }: BrandingUpdateRequest): Promise<Branding> {
    return this._fetch({
      url: `/v1/branding`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async delete(): Promise<object> {
    return this._fetch({
      url: `/v1/branding`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
