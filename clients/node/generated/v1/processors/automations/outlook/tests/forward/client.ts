import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIMailboxIdSub from "./mailboxId/client";

export default class APIForward extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  mailboxId = new APIMailboxIdSub(this._client);

}
