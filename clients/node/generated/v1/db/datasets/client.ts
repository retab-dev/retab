import { AbstractClient, CompositionClient } from '@/client';
import APISchemaIdSub from "./schemaId/client";
import APIAnnotationsJsonlSub from "./annotationsJsonl/client";

export default class APIDatasets extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaIdSub(this._client);
  annotationsJsonl = new APIAnnotationsJsonlSub(this._client);

}
