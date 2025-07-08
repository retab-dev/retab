import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';

export default class APIFields extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<object> {
    let res = await this._fetch({
      url: `/v1/extractions/fields`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
