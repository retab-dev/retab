import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIProcessorIdSub from "./processorId/client";

export default class APIProcessor extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  processorId = new APIProcessorIdSub(this._client);

}
