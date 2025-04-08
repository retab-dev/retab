import { AbstractClient, CompositionClient } from '@/client';
import { WebhookRequest } from "@/types";

export default class APIWebhookCloudflareMichael extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ spreadsheetId, organizationId, listFieldName, colorByLikelihood, appendLikelihoods, ...body }: { spreadsheetId: string, organizationId: string, listFieldName?: string | null, colorByLikelihood?: boolean, appendLikelihoods?: boolean } & WebhookRequest): Promise<object> {
    return this._fetch({
      url: `/v1/integrations/google_sheets/webhook_cloudflare_michael`,
      method: "POST",
      params: { "spreadsheet_id": spreadsheetId, "organization_id": organizationId, "list_field_name": listFieldName, "color_by_likelihood": colorByLikelihood, "append_likelihoods": appendLikelihoods },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
