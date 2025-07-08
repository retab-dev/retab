import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZWebhookRequest, WebhookRequest } from "@/types";

export default class APIWebhook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ spreadsheetId, worksheetId, organizationId, automationId, colorByLikelihood, ...body }: { spreadsheetId: string, worksheetId?: string, organizationId: string, automationId: string, colorByLikelihood?: boolean } & WebhookRequest): Promise<object> {
    let res = await this._fetch({
      url: `/v1/integrations/google_sheets/webhook`,
      method: "POST",
      params: { "spreadsheet_id": spreadsheetId, "worksheet_id": worksheetId, "organization_id": organizationId, "automation_id": automationId, "color_by_likelihood": colorByLikelihood },
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return z.object({}).parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
