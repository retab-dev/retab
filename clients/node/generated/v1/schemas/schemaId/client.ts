import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIExtractSub from "./extract/client";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  extract = new APIExtractSub(this._client);

}
