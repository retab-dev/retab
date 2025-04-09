import { AbstractClient, CompositionClient } from '@/client';
import APICreateCheckoutSessionSub from "./createCheckoutSession/client";
import APICreatePortalSessionSub from "./createPortalSession/client";
import APISubscriptionStatusSub from "./subscriptionStatus/client";
import APIWebhookSub from "./webhook/client";

export default class APIPayments extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  createCheckoutSession = new APICreateCheckoutSessionSub(this._client);
  createPortalSession = new APICreatePortalSessionSub(this._client);
  subscriptionStatus = new APISubscriptionStatusSub(this._client);
  webhook = new APIWebhookSub(this._client);

}
