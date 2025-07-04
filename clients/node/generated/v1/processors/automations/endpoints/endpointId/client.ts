import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { EndpointOutput, UpdateEndpointRequest } from "@/types";

export default class APIEndpointId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(endpointId: string): Promise<EndpointOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/endpoints/${endpointId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async put(endpointId: string, { ...body }: UpdateEndpointRequest): Promise<EndpointOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/endpoints/${endpointId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async delete(endpointId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/processors/automations/endpoints/${endpointId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
