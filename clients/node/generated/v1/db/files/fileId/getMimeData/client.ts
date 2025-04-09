import { AbstractClient, CompositionClient } from '@/client';
import { UiformTypesMimeMIMEData } from "@/types";

export default class APIGetMimeData extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(fileId: string): Promise<UiformTypesMimeMIMEData> {
    return this._fetch({
      url: `/v1/db/files/${fileId}/get_mime_data`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
