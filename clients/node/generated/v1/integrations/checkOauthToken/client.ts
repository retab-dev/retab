import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIApplicationNameSub from "./applicationName/client";

export default class APICheckOauthToken extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  applicationName = new APIApplicationNameSub(this._client);

}
