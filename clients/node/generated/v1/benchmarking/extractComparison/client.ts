import { AbstractClient, CompositionClient } from '@/client';
import { ComparisonRequest, ComparisonResponse } from "@/types";

export default class APIExtractComparison extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: ComparisonRequest): Promise<ComparisonResponse> {
    return this._fetch({
      url: `/v1/benchmarking/extract-comparison`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
