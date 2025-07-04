import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APISchemaIdSub from "./schemaId/client";

export default class APISchema extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaIdSub(this._client);

}
