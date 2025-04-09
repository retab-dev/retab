import { AbstractClient, CompositionClient } from '@/client';
import { AutomationDecisionRequest, AutomationDecisionResponse } from "@/types";

export default class APIAutomationDecision extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: AutomationDecisionRequest): Promise<AutomationDecisionResponse> {
    return this._fetch({
      url: `/v1/automations/outlook/vector_search/automation_decision`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
