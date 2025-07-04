import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { ExportToCsvRequest, MainServerServicesV1IntegrationsGoogleSheetsRoutesExportToCsvResponse } from "@/types";

export default class APIExportToCsv extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: ExportToCsvRequest): Promise<MainServerServicesV1IntegrationsGoogleSheetsRoutesExportToCsvResponse> {
    let res = await this._fetch({
      url: `/v1/integrations/google_sheets/export-to-csv`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
