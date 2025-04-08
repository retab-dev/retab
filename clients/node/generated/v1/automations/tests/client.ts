import { AbstractClient, CompositionClient } from '@/client';
import APIUpload from "./upload/client";
import APIWebhook from "./webhook/client";

export default class APITests extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  upload = new APIUpload(this);
  webhook = new APIWebhook(this);

}
