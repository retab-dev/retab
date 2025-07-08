import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZEnhanceSchemaRequest, EnhanceSchemaRequest, ZSchema, Schema } from "@/types";

export default class APIEnhance extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: EnhanceSchemaRequest): Promise<Schema> {
    let res = await this._fetch({
      url: `/v1/schemas/enhance`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZSchema.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
