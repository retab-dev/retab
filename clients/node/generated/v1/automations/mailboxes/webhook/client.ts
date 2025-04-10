import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIWebhook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ idempotencyKey }: { idempotencyKey?: string | null } = {}): Promise<any> {
    let res = await this._fetch({
      url: `/v1/automations/mailboxes/webhook`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      auth: ["HTTPBasic"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
