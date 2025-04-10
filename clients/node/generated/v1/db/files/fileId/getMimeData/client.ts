import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { UiformTypesMimeMIMEData } from "@/types";

export default class APIGetMimeData extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(fileId: string): Promise<UiformTypesMimeMIMEData> {
    let res = await this._fetch({
      url: `/v1/db/files/${fileId}/get_mime_data`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
