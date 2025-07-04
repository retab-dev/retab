import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APISubmitSub from "./submit/client";
import { OutlookOutput, UpdateOutlookRequest, OutlookOutput } from "@/types";

export default class APIOutlookPluginId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  submit = new APISubmitSub(this._client);

  async get(outlookPluginId: string): Promise<OutlookOutput> {
    let res = await this._fetch({
      url: `/v1/automations/outlook/${outlookPluginId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async put(outlookPluginId: string, { ...body }: UpdateOutlookRequest): Promise<OutlookOutput> {
    let res = await this._fetch({
      url: `/v1/automations/outlook/${outlookPluginId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async delete(outlookPluginId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/automations/outlook/${outlookPluginId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
