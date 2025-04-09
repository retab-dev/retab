import { AbstractClient, CompositionClient } from '@/client';
import APIMailboxIdSub from "./mailboxId/client";

export default class APIFromId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  mailboxId = new APIMailboxIdSub(this._client);

}
