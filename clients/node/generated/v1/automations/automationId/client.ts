import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { LinkOutput, MailboxOutput, EndpointOutput, OutlookOutput, UpdateLinkRequest, UpdateMailboxRequest, UpdateEndpointRequest, UpdateOutlookRequest, LinkOutput, MailboxOutput, EndpointOutput, OutlookOutput } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(automationId: string): Promise<LinkOutput | MailboxOutput | EndpointOutput | OutlookOutput> {
    let res = await this._fetch({
      url: `/v1/automations/${automationId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async put(automationId: string, { ...body }: UpdateLinkRequest | UpdateMailboxRequest | UpdateEndpointRequest | UpdateOutlookRequest): Promise<LinkOutput | MailboxOutput | EndpointOutput | OutlookOutput> {
    let res = await this._fetch({
      url: `/v1/automations/${automationId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
