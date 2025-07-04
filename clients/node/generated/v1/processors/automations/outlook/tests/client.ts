import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIForwardSub from "./forward/client";
import APIProcessSub from "./process/client";
import APIWebhookSub from "./webhook/client";
import APIEmailForwardingDefaultSub from "./emailForwardingDefault/client";

export default class APITests extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  forward = new APIForwardSub(this._client);
  process = new APIProcessSub(this._client);
  webhook = new APIWebhookSub(this._client);
  emailForwardingDefault = new APIEmailForwardingDefaultSub(this._client);

}
