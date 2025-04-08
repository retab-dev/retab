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
      params: {  },
      headers: {  },
    });
  }
  
  async put(endpointId: string, { ...body }: UpdateEndpointRequest): Promise<EndpointOutput> {
    return this._fetch({
      url: `/v1/automations/endpoints/${endpointId}`,
      method: "PUT",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
  async delete(endpointId: string): Promise<object> {
    return this._fetch({
      url: `/v1/automations/endpoints/${endpointId}`,
      method: "DELETE",
      params: {  },
      headers: {  },
    });
  }
  
}
