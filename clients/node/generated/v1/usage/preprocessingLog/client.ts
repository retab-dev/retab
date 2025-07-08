import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIExtractionIdSub from "./extractionId/client";

export default class APIPreprocessingLog extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  extractionId = new APIExtractionIdSub(this._client);

}
