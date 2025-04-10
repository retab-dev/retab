import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { IterationInput, IterationOutput } from "@/types";

export default class APIIterations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(experimentId: string, { iterationIndex, ...body }: { iterationIndex?: number | null } & IterationInput): Promise<IterationOutput> {
    let res = await this._fetch({
      url: `/v1/experiments/${experimentId}/iterations`,
      method: "POST",
      params: { "iteration_index": iterationIndex },
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
