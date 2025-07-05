import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIApplicationNameSub from "./applicationName/client";

export default class APICheckOauthToken extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  applicationName = new APIApplicationNameSub(this._client);

}
