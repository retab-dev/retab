import { AbstractClient, CompositionClient } from '@/client';
import APISchemasSub from "./schemas/client";

export default class APIOrganizations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemas = new APISchemasSub(this._client);

}
