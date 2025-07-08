import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APICreditsSub from "./credits/client";
import { ZAmount, Amount } from "@/types";

export default class APIProcessorId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  credits = new APICreditsSub(this._client);

  async get(processorId: string, { startDate, endDate }: { startDate?: Date | null, endDate?: Date | null } = {}): Promise<Amount> {
    let res = await this._fetch({
      url: `/v1/usage/processor/${processorId}`,
      method: "GET",
      params: { "start_date": startDate, "end_date": endDate },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZAmount.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
