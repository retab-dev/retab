import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIDistancesSub from "./distances/client";
import { ZRetabTypesEvalsIterationOutput, RetabTypesEvalsIterationOutput } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return ZRetabTypesEvalsIterationOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async delete(iterationId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/evals/iterations/${iterationId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
