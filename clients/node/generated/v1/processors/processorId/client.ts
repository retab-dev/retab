import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APISubmitSub from "./submit/client";
import { UpdateProcessorRequest, StoredProcessor } from "@/types";

export default class APIProcessorId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  submit = new APISubmitSub(this._client);

  async put(processorId: string, { ...body }: UpdateProcessorRequest): Promise<StoredProcessor> {
    let res = await this._fetch({
      url: `/v1/processors/${processorId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async get(processorId: string): Promise<StoredProcessor> {
    let res = await this._fetch({
      url: `/v1/processors/${processorId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async delete(processorId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/processors/${processorId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
