import { AbstractClient, CompositionClient } from '@/client';

export default class APISampleDocument extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(templateId: string): Promise<any> {
    return this._fetch({
      url: `/v1/schemas/history_templates/${templateId}/sample-document`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
