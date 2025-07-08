import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZComparisonRequest, ComparisonRequest, ZComparisonResponse, ComparisonResponse } from "@/types";

export default class APIExtractComparison extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: ComparisonRequest): Promise<ComparisonResponse> {
    let res = await this._fetch({
      url: `/v1/benchmarking/extract-comparison`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return ZComparisonResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
