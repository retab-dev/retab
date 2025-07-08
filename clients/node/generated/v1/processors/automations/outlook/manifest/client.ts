import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIIdSub from "./id/client";

export default class APIManifest extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  id = new APIIdSub(this._client);

}
