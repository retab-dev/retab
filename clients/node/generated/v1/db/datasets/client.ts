import { AbstractClient, CompositionClient } from '@/client';
import APISchemaId from "./schemaId/client";
import APIAnnotationsJsonl from "./annotationsJsonl/client";

export default class APIDatasets extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaId(this);
  annotationsJsonl = new APIAnnotationsJsonl(this);

}
