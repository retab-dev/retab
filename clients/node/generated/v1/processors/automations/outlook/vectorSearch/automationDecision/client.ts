import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { AutomationDecisionRequest, AutomationDecisionResponse } from "@/types";

export default class APIAutomationDecision extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: AutomationDecisionRequest): Promise<AutomationDecisionResponse> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook/vector_search/automation_decision`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
