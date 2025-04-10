import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APISchemasSub from "./schemas/client";

export default class APIAnalytics extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemas = new APISchemasSub(this._client);

}
