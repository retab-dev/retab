import { AbstractClient, CompositionClient } from '@/client';
import APISubmit from "./submit/client";
import { OutlookOutput, UpdateOutlookRequest, OutlookOutput } from "@/types";

export default class APIOutlookPluginId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  submit = new APISubmit(this);

  async get(outlookPluginId: string): Promise<OutlookOutput> {
    return this._fetch({
      url: `/v1/automations/outlook/${outlookPluginId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
  async put(outlookPluginId: string, { ...body }: UpdateOutlookRequest): Promise<OutlookOutput> {
    return this._fetch({
      url: `/v1/automations/outlook/${outlookPluginId}`,
      method: "PUT",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
  async delete(outlookPluginId: string): Promise<object> {
    return this._fetch({
      url: `/v1/automations/outlook/${outlookPluginId}`,
      method: "DELETE",
      params: {  },
      headers: {  },
    });
  }
  
}
