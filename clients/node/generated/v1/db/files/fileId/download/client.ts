import { AbstractClient, CompositionClient } from '@/client';

export default class APIDownload extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(fileId: string): Promise<any> {
    return this._fetch({
      url: `/v1/db/files/${fileId}/download`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
