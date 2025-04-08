import { AbstractClient, CompositionClient } from '@/client';
import { FileLink } from "@/types";

export default class APIDownloadLinkGcsPath extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ gcsPath }: { gcsPath: string }): Promise<FileLink> {
    return this._fetch({
      url: `/v1/db/files/download-link-gcs-path`,
      method: "GET",
      params: { "gcs_path": gcsPath },
      headers: {  },
    });
  }
  
}
