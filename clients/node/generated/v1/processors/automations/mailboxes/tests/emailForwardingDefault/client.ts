import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIMailboxIdSub from "./mailboxId/client";

export default class APIEmailForwardingDefault extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  mailboxId = new APIMailboxIdSub(this._client);

}
