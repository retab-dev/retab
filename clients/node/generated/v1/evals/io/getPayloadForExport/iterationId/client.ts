import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { RetabTypesEvalsEvaluationInput, MainServerServicesV1EvalsIoRoutesExportToCsvResponse } from "@/types";

export default class APIIterationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(iterationId: string, { delimiter, lineDelimiter, quote, ...body }: { delimiter?: string, lineDelimiter?: string, quote?: string } & RetabTypesEvalsEvaluationInput): Promise<MainServerServicesV1EvalsIoRoutesExportToCsvResponse> {
    let res = await this._fetch({
      url: `/v1/evals/io/get_payload_for_export/${iterationId}`,
      method: "POST",
      params: { "delimiter": delimiter, "line_delimiter": lineDelimiter, "quote": quote },
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
