import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { FileLink } from "@/types";

export default class APIDownloadLink extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(fileId: string): Promise<FileLink> {
    let res = await this._fetch({
      url: `/v1/db/files/${fileId}/download-link`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
