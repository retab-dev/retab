import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDocumentIdSub from "./documentId/client";

export default class APIDocuments extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  documentId = new APIDocumentIdSub(this._client);

}
