import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APISchemaIdSub from "./schemaId/client";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaIdSub(this._client);

}
