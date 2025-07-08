import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIV1Sub from "./v1/client";

export default class APIGenerated extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  v1 = new APIV1Sub(this._client);

}
