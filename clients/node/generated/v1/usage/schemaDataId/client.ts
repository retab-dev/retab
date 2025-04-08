import { AbstractClient, CompositionClient } from '@/client';
import APISchemaDataId from "./schemaDataId/client";

export default class APISchemaDataId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaDataId = new APISchemaDataId(this);

}
