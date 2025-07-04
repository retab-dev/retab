import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { FileLink } from "@/types";

export default class APIDownloadLinkGcsPath extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ gcsPath }: { gcsPath: string }): Promise<FileLink> {
    let res = await this._fetch({
      url: `/v1/db/files/download-link-gcs-path`,
      method: "GET",
      params: { "gcs_path": gcsPath },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
