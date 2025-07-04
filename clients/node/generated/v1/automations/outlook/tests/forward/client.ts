import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIEmailSub from "./email/client";

export default class APIForward extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  email = new APIEmailSub(this._client);

}
