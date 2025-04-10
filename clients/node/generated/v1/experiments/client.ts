import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIExperimentIdSub from "./experimentId/client";
import { ExperimentCreateRequest, Experiment } from "@/types";

export default class APIExperiments extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  experimentId = new APIExperimentIdSub(this._client);

  async post({ ...body }: ExperimentCreateRequest): Promise<Experiment> {
    let res = await this._fetch({
      url: `/v1/experiments`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
