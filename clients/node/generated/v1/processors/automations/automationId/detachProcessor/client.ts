import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZDetachProcessorRequest, DetachProcessorRequest, ZLinkOutput, LinkOutput, ZMailboxOutput, MailboxOutput, ZEndpointOutput, EndpointOutput, ZOutlookOutput, OutlookOutput } from "@/types";

export default class APIDetachProcessor extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(automationId: string, { ...body }: DetachProcessorRequest): Promise<LinkOutput | MailboxOutput | EndpointOutput | OutlookOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/${automationId}/detach-processor`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return z.union([ZLinkOutput, ZMailboxOutput, ZEndpointOutput, ZOutlookOutput]).parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
