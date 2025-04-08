import { AbstractClient, CompositionClient } from '@/client';

export default class APIWebhook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ IdempotencyKey }: { IdempotencyKey?: string | null }): Promise<any> {
    return this._fetch({
      url: `/v1/automations/mailboxes/webhook`,
      method: "POST",
      params: {  },
      headers: { "Idempotency-Key": IdempotencyKey },
    });
  }
  
}
