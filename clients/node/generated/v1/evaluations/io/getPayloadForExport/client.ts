import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIIterationIdSub from "./iterationId/client";

export default class APIGetPayloadForExport extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  iterationId = new APIIterationIdSub(this._client);

}
