import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDocumentIdSub from "./documentId/client";

export default class APIDistances extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  documentId = new APIDocumentIdSub(this._client);

}
