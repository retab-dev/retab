import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIReviewExtractionSub from "./reviewExtraction/client";
import APIDetachProcessorSub from "./detachProcessor/client";
import { ZLinkOutput, LinkOutput, ZMailboxOutput, MailboxOutput, ZEndpointOutput, EndpointOutput, ZOutlookOutput, OutlookOutput, ZUpdateLinkRequest, UpdateLinkRequest, ZUpdateMailboxRequest, UpdateMailboxRequest, ZUpdateEndpointRequest, UpdateEndpointRequest, ZUpdateOutlookRequest, UpdateOutlookRequest } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return z.union([ZLinkOutput, ZMailboxOutput, ZEndpointOutput, ZOutlookOutput]).parse(await res.json());
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
    if (res.headers.get("Content-Type") === "application/json") return z.union([ZLinkOutput, ZMailboxOutput, ZEndpointOutput, ZOutlookOutput]).parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async delete(automationId: string): Promise<any> {
    let res = await this._fetch({
      url: `/v1/processors/automations/${automationId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.any().parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
