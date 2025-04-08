import { AbstractClient, CompositionClient } from '@/client';
import APISchemas from "./schemas/client";

export default class APIOrganizations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemas = new APISchemas(this);

}
