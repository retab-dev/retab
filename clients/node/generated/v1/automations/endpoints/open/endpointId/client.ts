import { AbstractClient, CompositionClient } from '@/client';
import { EndpointOutput } from "@/types";

export default class APIEndpointId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(endpointId: string): Promise<EndpointOutput> {
    return this._fetch({
      url: `/v1/automations/endpoints/open/${endpointId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
