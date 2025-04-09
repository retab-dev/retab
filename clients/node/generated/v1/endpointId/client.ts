import { AbstractClient, CompositionClient } from '@/client';
import { BodyHandleEndpointProcessingV1EndpointIdPost, AutomationLog } from "@/types";

export default class APIEndpointId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(endpointId: string, { idempotencyKey, ...body }: { idempotencyKey?: string | null } & BodyHandleEndpointProcessingV1EndpointIdPost): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/${endpointId}`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "multipart/form-data",
    });
  }
  
}
