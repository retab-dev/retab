import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { DetachProcessorRequest, LinkOutput, MailboxOutput, EndpointOutput, OutlookOutput } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
