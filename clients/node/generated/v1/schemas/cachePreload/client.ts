import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZCachePreloadRequest, CachePreloadRequest } from "@/types";

export default class APICachePreload extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: CachePreloadRequest): Promise<object> {
    let res = await this._fetch({
      url: `/v1/schemas/cache_preload`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
