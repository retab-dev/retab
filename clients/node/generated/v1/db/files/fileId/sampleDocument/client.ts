import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APISampleDocument extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(fileId: string): Promise<any> {
    let res = await this._fetch({
      url: `/v1/db/files/${fileId}/sample-document`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
