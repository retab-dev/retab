import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';

export default class APITestIngestCompletion extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(): Promise<object> {
    let res = await this._fetch({
      url: `/v1/test_ingest_completion`,
      method: "POST",
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
