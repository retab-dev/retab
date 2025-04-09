import { AbstractClient, CompositionClient } from '@/client';
import APIDownloadLinkGcsPathSub from "./downloadLinkGcsPath/client";
import APIBatchSub from "./batch/client";
import APIFileIdSub from "./fileId/client";
import { ListFiles, BodyCreateFileV1DbFilesPost, DBFile } from "@/types";

export default class APIFiles extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  downloadLinkGcsPath = new APIDownloadLinkGcsPathSub(this._client);
  batch = new APIBatchSub(this._client);
  fileId = new APIFileIdSub(this._client);

  async get({ before, after, limit, order, filename, mimeType, includeEmbeddings, sortBy }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", filename?: string | null, mimeType?: string | null, includeEmbeddings?: boolean, sortBy?: string } = {}): Promise<ListFiles> {
    return this._fetch({
      url: `/v1/db/files`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "filename": filename, "mime_type": mimeType, "include_embeddings": includeEmbeddings, "sort_by": sortBy },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async post({ ...body }: BodyCreateFileV1DbFilesPost): Promise<DBFile> {
    return this._fetch({
      url: `/v1/db/files`,
      method: "POST",
      body: body,
      bodyMime: "multipart/form-data",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
