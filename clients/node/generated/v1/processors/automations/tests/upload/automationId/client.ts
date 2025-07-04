import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyTestDocumentUploadV1ProcessorsAutomationsTestsUploadAutomationIdPost, AutomationLog } from "@/types";

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
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
