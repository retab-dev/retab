import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIMailboxIdSub from "./mailboxId/client";

export default class APIWebhook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  mailboxId = new APIMailboxIdSub(this._client);

}
