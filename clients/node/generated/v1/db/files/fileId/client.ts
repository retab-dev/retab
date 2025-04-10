import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDownloadLinkSub from "./downloadLink/client";
import APISampleDocumentSub from "./sampleDocument/client";
import APIDownloadSub from "./download/client";
import APIGetMimeDataSub from "./getMimeData/client";
import { DBFile } from "@/types";

export default class APIFileId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  downloadLink = new APIDownloadLinkSub(this._client);
  sampleDocument = new APISampleDocumentSub(this._client);
  download = new APIDownloadSub(this._client);
  getMimeData = new APIGetMimeDataSub(this._client);

  async get(fileId: string): Promise<DBFile> {
    let res = await this._fetch({
      url: `/v1/db/files/${fileId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
