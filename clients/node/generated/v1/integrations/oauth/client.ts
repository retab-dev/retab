import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIGoogleSub from "./google/client";

export default class APIOauth extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  google = new APIGoogleSub(this._client);

}
