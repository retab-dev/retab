import { AbstractClient, CompositionClient } from '@/client';
import APISchemaId from "./schemaId/client";

export default class APIUsageschemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaId(this);

}
