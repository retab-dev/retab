import { AbstractClient, CompositionClient } from '@/client';

export default class APIWebhook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey }: { idempotencyKey?: string | null } = {}): Promise<any> {
    return this._fetch({
      url: `/v1/automations/mailboxes/webhook`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      auth: ["HTTPBasic"],
    });
  }
  
}
