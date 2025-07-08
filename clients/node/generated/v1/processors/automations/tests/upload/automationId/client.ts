import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZBodyTestDocumentUploadV1ProcessorsAutomationsTestsUploadAutomationIdPost, BodyTestDocumentUploadV1ProcessorsAutomationsTestsUploadAutomationIdPost, ZAutomationLog, AutomationLog } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(automationId: string, { ...body }: BodyTestDocumentUploadV1ProcessorsAutomationsTestsUploadAutomationIdPost): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/processors/automations/tests/upload/${automationId}`,
      method: "POST",
      body: body,
      bodyMime: "multipart/form-data",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZAutomationLog.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
