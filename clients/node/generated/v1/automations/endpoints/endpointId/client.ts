import { AbstractClient, CompositionClient } from '@/client';
import { EndpointOutput, UpdateEndpointRequest, EndpointOutput } from "@/types";

export default class APIEndpointId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(endpointId: string): Promise<EndpointOutput> {
    return this._fetch({
      url: `/v1/automations/endpoints/${endpointId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async put(endpointId: string, { ...body }: UpdateEndpointRequest): Promise<EndpointOutput> {
    return this._fetch({
      url: `/v1/automations/endpoints/${endpointId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async delete(endpointId: string): Promise<object> {
    return this._fetch({
      url: `/v1/automations/endpoints/${endpointId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
