import { AbstractClient, CompositionClient } from '@/client';
import APIEmailSub from "./email/client";

export default class APIEmailForwardingDefault extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  email = new APIEmailSub(this._client);

}
