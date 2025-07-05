import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDistancesSub from "./distances/client";
import { RetabTypesEvalsIterationOutput } from "@/types";

export default class APIIterationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  distances = new APIDistancesSub(this._client);

  async get(iterationId: string): Promise<RetabTypesEvalsIterationOutput> {
    let res = await this._fetch({
      url: `/v1/evals/iterations/${iterationId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async delete(iterationId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/evals/iterations/${iterationId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
