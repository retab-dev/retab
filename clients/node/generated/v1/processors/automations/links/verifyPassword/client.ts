import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APILinkIdSub from "./linkId/client";

export default class APIVerifyPassword extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  linkId = new APILinkIdSub(this._client);

}
