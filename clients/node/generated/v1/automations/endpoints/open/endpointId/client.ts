import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { EndpointOutput } from "@/types";

export default class APIEndpointId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(endpointId: string): Promise<EndpointOutput> {
    let res = await this._fetch({
      url: `/v1/automations/endpoints/open/${endpointId}`,
      method: "GET",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
