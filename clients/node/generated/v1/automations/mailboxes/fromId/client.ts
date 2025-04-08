import { AbstractClient, CompositionClient } from '@/client';
import APIMailboxId from "./mailboxId/client";

export default class APIFromId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  mailboxId = new APIMailboxId(this);

}
