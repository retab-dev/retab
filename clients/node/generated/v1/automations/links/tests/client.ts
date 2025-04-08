import { AbstractClient, CompositionClient } from '@/client';
import APIReception from "./reception/client";
import APIUpload from "./upload/client";
import APIWebhook from "./webhook/client";

export default class APITests extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  reception = new APIReception(this);
  upload = new APIUpload(this);
  webhook = new APIWebhook(this);

}
