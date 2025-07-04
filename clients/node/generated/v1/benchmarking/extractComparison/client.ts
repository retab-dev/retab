import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ComparisonRequest, ComparisonResponse } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
