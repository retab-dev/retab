import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIExportDocuments extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(evaluationId: string): Promise<object> {
    let res = await this._fetch({
      url: `/v1/evals/io/${evaluationId}/export_documents`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
