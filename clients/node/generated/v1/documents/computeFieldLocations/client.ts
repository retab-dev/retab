import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZComputeFieldLocationsRequest, ComputeFieldLocationsRequest, ZComputeFieldLocationsResponse, ComputeFieldLocationsResponse } from "@/types";

export default class APIComputeFieldLocations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: ComputeFieldLocationsRequest): Promise<ComputeFieldLocationsResponse> {
    let res = await this._fetch({
      url: `/v1/documents/compute_field_locations`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZComputeFieldLocationsResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
