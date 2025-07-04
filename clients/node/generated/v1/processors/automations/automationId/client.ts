import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIReviewExtractionSub from "./reviewExtraction/client";
import APIDetachProcessorSub from "./detachProcessor/client";
import { LinkOutput, MailboxOutput, EndpointOutput, OutlookOutput, UpdateLinkRequest, UpdateMailboxRequest, UpdateEndpointRequest, UpdateOutlookRequest } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  reviewExtraction = new APIReviewExtractionSub(this._client);
  detachProcessor = new APIDetachProcessorSub(this._client);

  async get(automationId: string): Promise<LinkOutput | MailboxOutput | EndpointOutput | OutlookOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/${automationId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async put(automationId: string, { ...body }: UpdateLinkRequest | UpdateMailboxRequest | UpdateEndpointRequest | UpdateOutlookRequest): Promise<LinkOutput | MailboxOutput | EndpointOutput | OutlookOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/${automationId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async delete(automationId: string): Promise<any> {
    let res = await this._fetch({
      url: `/v1/processors/automations/${automationId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
