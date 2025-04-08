import { AbstractClient, CompositionClient } from '@/client';
import APIAutomations from "./automations/client";
import { Branding, BrandingUpdateRequest, Branding } from "@/types";

export default class APIBranding extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  automations = new APIAutomations(this);

  async get(): Promise<Branding> {
    return this._fetch({
      url: `/v1/branding`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
  async post({ ...body }: BrandingUpdateRequest): Promise<Branding> {
    return this._fetch({
      url: `/v1/branding`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
  async delete(): Promise<object> {
    return this._fetch({
      url: `/v1/branding`,
      method: "DELETE",
      params: {  },
      headers: {  },
    });
  }
  
}
