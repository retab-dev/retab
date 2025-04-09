import { AbstractClient, CompositionClient } from '@/client';

export default class APISampleDocument extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(extractionId: string): Promise<any> {
    return this._fetch({
      url: `/v1/extractions_logs/${extractionId}/sample-document`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
