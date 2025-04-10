import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIIterationsSub from "./iterations/client";
import { Experiment, ExperimentCreateRequest, Experiment } from "@/types";

export default class APIExperimentId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  iterations = new APIIterationsSub(this._client);

  async get(experimentId: string): Promise<Experiment> {
    let res = await this._fetch({
      url: `/v1/experiments/${experimentId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async put(experimentId: string, { ...body }: ExperimentCreateRequest): Promise<Experiment> {
    let res = await this._fetch({
      url: `/v1/experiments/${experimentId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
