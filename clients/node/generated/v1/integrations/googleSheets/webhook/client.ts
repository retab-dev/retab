import { AbstractClient, CompositionClient } from '@/client';
import { WebhookRequest } from "@/types";

export default class APIWebhook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ spreadsheetId, sheetId, organizationId, listFieldName, colorByLikelihood, appendLikelihoods, ...body }: { spreadsheetId: string, sheetId?: string, organizationId: string, listFieldName?: string | null, colorByLikelihood?: boolean, appendLikelihoods?: boolean } & WebhookRequest): Promise<object> {
    return this._fetch({
      url: `/v1/integrations/google_sheets/webhook`,
      method: "POST",
      params: { "spreadsheet_id": spreadsheetId, "sheet_id": sheetId, "organization_id": organizationId, "list_field_name": listFieldName, "color_by_likelihood": colorByLikelihood, "append_likelihoods": appendLikelihoods },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
