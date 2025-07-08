import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZAutomationLog, AutomationLog } from "@/types";

export default class APIMailboxId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(mailboxId: string): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/processors/automations/mailboxes/tests/webhook/${mailboxId}`,
      method: "POST",
    });
    if (res.headers.get("Content-Type") === "application/json") return ZAutomationLog.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
