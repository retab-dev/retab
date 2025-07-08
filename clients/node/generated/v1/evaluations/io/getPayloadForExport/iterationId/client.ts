import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZRetabTypesEvaluationsModelEvaluationInput, RetabTypesEvaluationsModelEvaluationInput, ZMainServerServicesV1EvaluationsIoRoutesExportToCsvResponse, MainServerServicesV1EvaluationsIoRoutesExportToCsvResponse } from "@/types";

export default class APIIterationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(iterationId: string, { delimiter, lineDelimiter, quote, ...body }: { delimiter?: string, lineDelimiter?: string, quote?: string } & RetabTypesEvaluationsModelEvaluationInput): Promise<MainServerServicesV1EvaluationsIoRoutesExportToCsvResponse> {
    let res = await this._fetch({
      url: `/v1/evaluations/io/get_payload_for_export/${iterationId}`,
      method: "POST",
      params: { "delimiter": delimiter, "line_delimiter": lineDelimiter, "quote": quote },
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return ZMainServerServicesV1EvaluationsIoRoutesExportToCsvResponse.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
