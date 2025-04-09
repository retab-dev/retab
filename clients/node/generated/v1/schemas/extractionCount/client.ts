import { AbstractClient, CompositionClient } from '@/client';
import APISchemaIdSub from "./schemaId/client";

export default class APIExtractionCount extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaIdSub(this._client);

}
