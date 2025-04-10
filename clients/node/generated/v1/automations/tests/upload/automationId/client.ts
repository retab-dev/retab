import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyTestDocumentUploadV1AutomationsTestsUploadAutomationIdPost, AutomationLog } from "@/types";

export default class APIAutomationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(automationId: string, { ...body }: BodyTestDocumentUploadV1AutomationsTestsUploadAutomationIdPost): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/automations/tests/upload/${automationId}`,
      method: "POST",
      body: body,
      bodyMime: "multipart/form-data",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
