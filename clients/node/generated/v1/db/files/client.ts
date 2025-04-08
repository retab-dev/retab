import { AbstractClient, CompositionClient } from '@/client';
import APIDownloadLinkGcsPath from "./downloadLinkGcsPath/client";
import APIBatch from "./batch/client";
import APIFileId from "./fileId/client";
import { ListFiles, BodyCreateFileV1DbFilesPost, DBFile } from "@/types";

export default class APIFiles extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  downloadLinkGcsPath = new APIDownloadLinkGcsPath(this);
  batch = new APIBatch(this);
  fileId = new APIFileId(this);

  async get({ before, after, limit, order, filename, mimeType, includeEmbeddings, sortBy }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", filename?: string | null, mimeType?: string | null, includeEmbeddings?: boolean, sortBy?: string }): Promise<ListFiles> {
    return this._fetch({
      url: `/v1/db/files`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "filename": filename, "mime_type": mimeType, "include_embeddings": includeEmbeddings, "sort_by": sortBy },
      headers: {  },
    });
  }
  
  async post({ ...body }: BodyCreateFileV1DbFilesPost): Promise<DBFile> {
    return this._fetch({
      url: `/v1/db/files`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "multipart/form-data",
    });
  }
  
}
