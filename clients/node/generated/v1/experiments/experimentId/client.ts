import { AbstractClient, CompositionClient } from '@/client';
import APIIterations from "./iterations/client";
import { Experiment } from "@/types";

export default class APIExperimentId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  iterations = new APIIterations(this);

  async get(experimentId: string): Promise<Experiment> {
    return this._fetch({
      url: `/v1/experiments/${experimentId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
