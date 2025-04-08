import { AbstractClient, CompositionClient } from '@/client';
import { FileLink } from "@/types";

export default class APIDownloadLink extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(fileId: string): Promise<FileLink> {
    return this._fetch({
      url: `/v1/db/files/${fileId}/download-link`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
