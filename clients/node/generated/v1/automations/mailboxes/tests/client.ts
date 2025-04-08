import { AbstractClient, CompositionClient } from '@/client';
import APIForward from "./forward/client";
import APIProcess from "./process/client";
import APIWebhook from "./webhook/client";
import APIEmailForwardingDefault from "./emailForwardingDefault/client";

export default class APITests extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  forward = new APIForward(this);
  process = new APIProcess(this);
  webhook = new APIWebhook(this);
  emailForwardingDefault = new APIEmailForwardingDefault(this);

}
