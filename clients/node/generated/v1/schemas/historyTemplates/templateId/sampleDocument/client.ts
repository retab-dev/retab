import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APISampleDocument extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(templateId: string): Promise<any> {
    let res = await this._fetch({
      url: `/v1/schemas/history_templates/${templateId}/sample-document`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
