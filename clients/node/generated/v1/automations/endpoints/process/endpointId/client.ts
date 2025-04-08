import { AbstractClient, CompositionClient } from '@/client';
import { BodyHandleEndpointProcessingV1AutomationsEndpointsProcessEndpointIdPost, AutomationLog } from "@/types";

export default class APIEndpointId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(endpointId: string, { IdempotencyKey, ...body }: { IdempotencyKey?: string | null } & BodyHandleEndpointProcessingV1AutomationsEndpointsProcessEndpointIdPost): Promise<AutomationLog> {
    return this._fetch({
      url: `/v1/automations/endpoints/process/${endpointId}`,
      method: "POST",
      params: {  },
      headers: { "Idempotency-Key": IdempotencyKey },
      body: body,
      bodyMime: "multipart/form-data",
    });
  }
  
}
