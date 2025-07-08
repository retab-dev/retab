import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZDocumentUploadRequest, DocumentUploadRequest, ZEmailDataOutput, EmailDataOutput } from "@/types";

export default class APIMailboxId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(mailboxId: string, { ...body }: DocumentUploadRequest): Promise<EmailDataOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/mailboxes/tests/forward/${mailboxId}`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZEmailDataOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
