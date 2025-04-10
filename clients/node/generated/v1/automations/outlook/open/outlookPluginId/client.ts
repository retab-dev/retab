import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { OutlookOutput } from "@/types";

export default class APIOutlookPluginId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(outlookPluginId: string): Promise<OutlookOutput> {
    let res = await this._fetch({
      url: `/v1/automations/outlook/open/${outlookPluginId}`,
      method: "GET",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
