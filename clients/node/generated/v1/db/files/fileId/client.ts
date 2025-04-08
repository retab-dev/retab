import { AbstractClient, CompositionClient } from '@/client';
import APIDownloadLink from "./downloadLink/client";
import APISampleDocument from "./sampleDocument/client";
import APIDownload from "./download/client";
import APIGetMimeData from "./getMimeData/client";
import { DBFile } from "@/types";

export default class APIFileId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  downloadLink = new APIDownloadLink(this);
  sampleDocument = new APISampleDocument(this);
  download = new APIDownload(this);
  getMimeData = new APIGetMimeData(this);

  async get(fileId: string): Promise<DBFile> {
    return this._fetch({
      url: `/v1/db/files/${fileId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
