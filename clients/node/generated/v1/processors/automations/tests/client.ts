import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIUploadSub from "./upload/client";
import APIWebhookSub from "./webhook/client";

export default class APITests extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  upload = new APIUploadSub(this._client);
  webhook = new APIWebhookSub(this._client);

}
