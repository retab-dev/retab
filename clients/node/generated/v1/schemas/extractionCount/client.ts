import { AbstractClient, CompositionClient } from '@/client';
import APISchemaId from "./schemaId/client";

export default class APIExtractionCount extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaId(this);

}
