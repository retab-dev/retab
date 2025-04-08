import { AbstractClient, CompositionClient } from '@/client';
import { FileScoreIndex } from "@/types";

export default class APIFileId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(fileId: string): Promise<FileScoreIndex> {
    return this._fetch({
      url: `/v1/automations/outlook/vector_search/file_scores/${fileId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
