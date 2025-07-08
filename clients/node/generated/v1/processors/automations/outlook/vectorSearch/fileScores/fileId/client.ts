import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZFileScoreIndex, FileScoreIndex } from "@/types";

export default class APIFileId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(fileId: string): Promise<FileScoreIndex> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook/vector_search/file_scores/${fileId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZFileScoreIndex.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
