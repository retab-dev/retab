import { AbstractClient, CompositionClient } from '@/client';
import { IterationInput, IterationOutput } from "@/types";

export default class APIIterations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(experimentId: string, { ...body }: IterationInput): Promise<IterationOutput> {
    return this._fetch({
      url: `/v1/experiments/${experimentId}/iterations`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
