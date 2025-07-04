import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIStreamSub from "./stream/client";
import { ProcessIterationDocument, RetabParsedChatCompletionOutput } from "@/types";

export default class APIProcess extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  stream = new APIStreamSub(this._client);

  async post(evaluationId: string, iterationId: string, documentId: string, { ...body }: ProcessIterationDocument): Promise<RetabParsedChatCompletionOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations/${iterationId}/documents/${documentId}/process`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
