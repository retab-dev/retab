import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIDocumentIdSub from "./documentId/client";

export default class APIDistances extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  documentId = new APIDocumentIdSub(this._client);

}
