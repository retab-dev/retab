import { AbstractClient, CompositionClient } from '@/client';
import APIExperimentId from "./experimentId/client";
import { ExperimentCreateRequest, Experiment } from "@/types";

export default class APIExperiments extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  experimentId = new APIExperimentId(this);

  async post({ ...body }: ExperimentCreateRequest): Promise<Experiment> {
    return this._fetch({
      url: `/v1/experiments`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
