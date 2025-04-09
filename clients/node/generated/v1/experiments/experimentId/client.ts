import { AbstractClient, CompositionClient } from '@/client';
import APIIterationsSub from "./iterations/client";
import { Experiment } from "@/types";

export default class APIExperimentId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  iterations = new APIIterationsSub(this._client);

  async get(experimentId: string): Promise<Experiment> {
    return this._fetch({
      url: `/v1/experiments/${experimentId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
