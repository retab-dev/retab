import { AbstractClient, CompositionClient } from '@/client';
import APIApplicationName from "./applicationName/client";

export default class APICheckOauthToken extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  applicationName = new APIApplicationName(this);

}
