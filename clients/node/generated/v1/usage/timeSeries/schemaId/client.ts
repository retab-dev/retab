import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APISchemaIdSub from "./schemaId/client";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaIdSub(this._client);

}
