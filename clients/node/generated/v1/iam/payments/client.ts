import { AbstractClient, CompositionClient } from '@/client';
import APICreateCheckoutSession from "./createCheckoutSession/client";
import APICreatePortalSession from "./createPortalSession/client";
import APISubscriptionStatus from "./subscriptionStatus/client";
import APIWebhook from "./webhook/client";

export default class APIPayments extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  createCheckoutSession = new APICreateCheckoutSession(this);
  createPortalSession = new APICreatePortalSession(this);
  subscriptionStatus = new APISubscriptionStatus(this);
  webhook = new APIWebhook(this);

}
