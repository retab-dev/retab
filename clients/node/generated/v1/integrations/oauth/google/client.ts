import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIGetAuthUrlSub from "./getAuthUrl/client";
import APIGetTokenSub from "./getToken/client";

export default class APIGoogle extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  getAuthUrl = new APIGetAuthUrlSub(this._client);
  getToken = new APIGetTokenSub(this._client);

}
