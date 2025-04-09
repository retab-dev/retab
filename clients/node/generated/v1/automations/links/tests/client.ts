import { AbstractClient, CompositionClient } from '@/client';
import APIReceptionSub from "./reception/client";
import APIUploadSub from "./upload/client";
import APIWebhookSub from "./webhook/client";

export default class APITests extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  reception = new APIReceptionSub(this._client);
  upload = new APIUploadSub(this._client);
  webhook = new APIWebhookSub(this._client);

}
