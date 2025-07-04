import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ProcessIterationDocument } from "@/types";

export default class APIStream extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(evaluationId: string, iterationId: string, documentId: string, { ...body }: ProcessIterationDocument): Promise<AsyncGenerator<string>> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations/${iterationId}/documents/${documentId}/process/stream`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/stream+json") return streamResponse(res);
    throw new Error("Bad content type");
  }
  
}
