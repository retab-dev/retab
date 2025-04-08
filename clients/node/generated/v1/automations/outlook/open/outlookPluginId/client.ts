import { AbstractClient, CompositionClient } from '@/client';
import { OutlookOutput } from "@/types";

export default class APIOutlookPluginId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(outlookPluginId: string): Promise<OutlookOutput> {
    return this._fetch({
      url: `/v1/automations/outlook/open/${outlookPluginId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
