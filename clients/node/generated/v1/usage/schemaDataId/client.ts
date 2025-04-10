import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APISchemaDataIdSub from "./schemaDataId/client";

export default class APISchemaDataId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaDataId = new APISchemaDataIdSub(this._client);

}
