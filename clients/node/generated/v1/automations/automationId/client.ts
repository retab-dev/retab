import { AbstractClient, CompositionClient } from '@/client';
import { LinkOutput, MailboxOutput, EndpointOutput, OutlookOutput, UpdateLinkRequest, UpdateMailboxRequest, UpdateEndpointRequest, UpdateOutlookRequest, LinkOutput, MailboxOutput, EndpointOutput, OutlookOutput } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(automationId: string): Promise<LinkOutput | MailboxOutput | EndpointOutput | OutlookOutput> {
    return this._fetch({
      url: `/v1/automations/${automationId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
  async put(automationId: string, { ...body }: UpdateLinkRequest | UpdateMailboxRequest | UpdateEndpointRequest | UpdateOutlookRequest): Promise<LinkOutput | MailboxOutput | EndpointOutput | OutlookOutput> {
    return this._fetch({
      url: `/v1/automations/${automationId}`,
      method: "PUT",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
